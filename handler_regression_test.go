package goapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSetExampleNilPointerDoesNotLoop(t *testing.T) {
	var example *string
	field := &paramField{
		meta: &paramMeta{
			example: example,
		},
	}
	val := reflect.New(reflect.TypeOf("")).Elem()
	done := make(chan struct{})

	go func() {
		_ = (&handler{}).setExample(val, field, false)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("setExample should return when example is a nil pointer")
	}
}

type contextLoggerRegression struct {
	ctx   *Context
	calls int
}

func (l *contextLoggerRegression) Debug(format string, a ...any) {}
func (l *contextLoggerRegression) Info(format string, a ...any)  {}
func (l *contextLoggerRegression) Warn(format string, a ...any)  {}
func (l *contextLoggerRegression) Error(format string, a ...any) {}
func (l *contextLoggerRegression) Fatal(format string, a ...any) {}
func (l *contextLoggerRegression) WithFields(keysAndValues ...any) Logger {
	return l
}

func (l *contextLoggerRegression) WithContext(ctx *Context) Logger {
	l.calls++
	return &contextLoggerRegression{ctx: ctx}
}

func TestHandleLoggerUsesLoggerWithContext(t *testing.T) {
	base := &contextLoggerRegression{}
	ctx := &Context{log: base}

	(&handlerServer{}).handleLogger(ctx)

	got, ok := ctx.log.(*contextLoggerRegression)
	if !ok {
		t.Fatalf("logger type = %T, want *contextLoggerRegression", ctx.log)
	}
	if got == base {
		t.Fatal("handleLogger should use the logger returned by WithContext")
	}
	if got.ctx != ctx {
		t.Fatal("WithContext logger should receive the request context")
	}
	if base.ctx != nil {
		t.Fatal("base logger should not be mutated with request context")
	}
	if base.calls != 1 {
		t.Fatalf("WithContext calls = %d, want 1", base.calls)
	}
}

func TestContextCopyRebindsLoggerWithContext(t *testing.T) {
	base := &contextLoggerRegression{}
	ctx := &Context{
		Request: httptest.NewRequest(http.MethodGet, "/", nil),
		log:     base,
		baseLog: base,
	}

	(&handlerServer{}).handleLogger(ctx)
	original, ok := ctx.log.(*contextLoggerRegression)
	if !ok {
		t.Fatalf("logger type = %T, want *contextLoggerRegression", ctx.log)
	}
	if original.ctx != ctx {
		t.Fatal("request logger should be bound to the original context")
	}

	cp := ctx.Copy()
	got, ok := cp.log.(*contextLoggerRegression)
	if !ok {
		t.Fatalf("copy logger type = %T, want *contextLoggerRegression", cp.log)
	}
	if cp.baseLog != base {
		t.Fatal("Copy should preserve the base logger")
	}
	if got == original {
		t.Fatal("Copy should not reuse the request-scoped logger")
	}
	if got.ctx != cp {
		t.Fatal("Copy should bind LoggerWithContext to the copied context")
	}
	if original.ctx != ctx {
		t.Fatal("Copy should not mutate the original request logger")
	}
	if base.calls != 2 {
		t.Fatalf("WithContext calls = %d, want 2", base.calls)
	}
}

type nilContextLoggerRegression struct {
	nopLogger
	calls int
}

func (l *nilContextLoggerRegression) WithContext(ctx *Context) Logger {
	l.calls++
	return nil
}

func TestLoggerWithContextNilKeepsBaseLogger(t *testing.T) {
	t.Run("request context", func(t *testing.T) {
		base := &nilContextLoggerRegression{}
		ctx := &Context{log: base}

		(&handlerServer{}).handleLogger(ctx)

		if ctx.log != base {
			t.Fatalf("logger = %T, want base logger", ctx.log)
		}
		if base.calls != 1 {
			t.Fatalf("WithContext calls = %d, want 1", base.calls)
		}
	})

	t.Run("copied context", func(t *testing.T) {
		base := &nilContextLoggerRegression{}
		ctx := &Context{
			Request: httptest.NewRequest(http.MethodGet, "/", nil),
			log:     base,
			baseLog: base,
		}

		cp := ctx.Copy()

		if cp.log != base {
			t.Fatalf("copy logger = %T, want base logger", cp.log)
		}
		if cp.baseLog != base {
			t.Fatal("Copy should preserve the base logger")
		}
		if base.calls != 1 {
			t.Fatalf("WithContext calls = %d, want 1", base.calls)
		}
	})
}

type marshalErrorResponse struct{}

func (marshalErrorResponse) GetStatus() int { return http.StatusCreated }
func (marshalErrorResponse) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal failed")
}

type statusBodyResponse struct {
	status int
	body   any
}

func (r statusBodyResponse) GetStatus() int { return r.status }
func (r statusBodyResponse) GetBody() any   { return r.body }

func TestHandleResponseMarshalErrorBeforeStatusWritten(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := &Context{
		Request:      req,
		log:          nopLogger{},
		Params:       &Params{},
		skippedNodes: &[]skippedNode{},
	}
	ctx.writermem.reset(rec)
	ctx.Writer = &ctx.writermem

	server := &handlerServer{
		handle: &handler{
			childMap: map[string]returnObjChild{
				"": {responseMediaTypes: []MediaType{JSON}},
			},
			errorMap: map[string]*errorInfo{
				"": {
					errorFunc: func(err error) any {
						return statusBodyResponse{
							status: http.StatusInternalServerError,
							body:   map[string]string{"error": err.Error()},
						}
					},
				},
			},
		},
	}

	server.handleResponse(ctx, marshalErrorResponse{})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status code: got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body should be JSON: %v", err)
	}
	if !strings.Contains(body["error"], "marshal failed") {
		t.Fatalf("error body: got %q", body["error"])
	}
}

func TestResponseWriterKeepsFirstStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	writer := &responseWriter{}
	writer.reset(rec)

	writer.WriteHeader(http.StatusCreated)
	writer.WriteHeader(http.StatusInternalServerError)
	if writer.Status() != http.StatusCreated {
		t.Fatalf("Status(): got %d want %d", writer.Status(), http.StatusCreated)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("recorder code: got %d want %d", rec.Code, http.StatusCreated)
	}

	rec = httptest.NewRecorder()
	writer.reset(rec)
	if _, err := writer.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	writer.WriteHeader(http.StatusInternalServerError)
	if writer.Status() != http.StatusOK {
		t.Fatalf("Status() after Write: got %d want %d", writer.Status(), http.StatusOK)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("recorder code after Write: got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestResponseWriterFlushCommitsStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	writer := &responseWriter{}
	writer.reset(rec)

	writer.Flush()
	writer.WriteHeader(http.StatusInternalServerError)

	if writer.Status() != http.StatusOK {
		t.Fatalf("Status() after Flush then WriteHeader: got %d want %d", writer.Status(), http.StatusOK)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("recorder code after Flush: got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestGenerateRequestIDPrefersXRequestIDHeaderWhenEnabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "req-from-upstream")
	ctx := &Context{Request: req}
	rec := httptest.NewRecorder()
	ctx.writermem.reset(rec)
	ctx.Writer = &ctx.writermem
	server := &handlerServer{
		handle: &handler{
			api: &API{
				GenerateRequestID:   true,
				UseXRequestIDHeader: true,
			},
		},
	}

	server.generateRequestID(ctx)

	if ctx.RequestID != "req-from-upstream" {
		t.Fatalf("RequestID: got %q want %q", ctx.RequestID, "req-from-upstream")
	}
	if got := rec.Header().Get("X-Request-ID"); got != "req-from-upstream" {
		t.Fatalf("X-Request-ID response header: got %q want %q", got, "req-from-upstream")
	}
}

func TestGenerateRequestIDWritesGeneratedXRequestIDHeaderWhenEnabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := &Context{Request: req}
	rec := httptest.NewRecorder()
	ctx.writermem.reset(rec)
	ctx.Writer = &ctx.writermem
	server := &handlerServer{
		handle: &handler{
			api: &API{
				GenerateRequestID:   true,
				UseXRequestIDHeader: true,
			},
		},
	}

	server.generateRequestID(ctx)

	if ctx.RequestID == "" {
		t.Fatal("RequestID should be generated")
	}
	if got := rec.Header().Get("X-Request-ID"); got != ctx.RequestID {
		t.Fatalf("X-Request-ID response header: got %q want %q", got, ctx.RequestID)
	}
}

func TestGenerateRequestIDIgnoresXRequestIDWhenDisabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "req-from-upstream")
	ctx := &Context{Request: req}
	rec := httptest.NewRecorder()
	ctx.writermem.reset(rec)
	ctx.Writer = &ctx.writermem
	server := &handlerServer{
		handle: &handler{
			api: &API{
				GenerateRequestID: true,
			},
		},
	}

	server.generateRequestID(ctx)

	if ctx.RequestID == "" {
		t.Fatal("RequestID should be generated")
	}
	if ctx.RequestID == "req-from-upstream" {
		t.Fatal("RequestID should not use X-Request-ID when UseXRequestIDHeader is false")
	}
	if got := rec.Header().Get("X-Request-ID"); got != "" {
		t.Fatalf("X-Request-ID response header should be empty, got %q", got)
	}
}

type failingHTTPWriter struct {
	header http.Header
	err    error
}

func (w *failingHTTPWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w *failingHTTPWriter) Write([]byte) (int, error) { return 0, w.err }
func (w *failingHTTPWriter) WriteHeader(int)           {}

func TestCopyReaderReturnsWriteError(t *testing.T) {
	writeErr := errors.New("write failed")
	writer := &responseWriter{}
	writer.reset(&failingHTTPWriter{err: writeErr})

	err := (&handlerServer{}).copyReader(writer, io.NopCloser(strings.NewReader("payload")))
	if !errors.Is(err, writeErr) {
		t.Fatalf("copyReader error: got %v want %v", err, writeErr)
	}
}

func TestNotFindMethodAllowedDoesNotReuseParamsFromFailedMatch(t *testing.T) {
	getRoot := &node{}
	if err := getRoot.addRoute("/users/{id}/details", fakeHandler); err != nil {
		t.Fatal(err)
	}
	postRoot := &node{}
	if err := postRoot.addRoute("/users/{id}", fakeHandler); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:      httptest.NewRequest(http.MethodGet, "/users/42", nil),
		log:          nopLogger{},
		Params:       &Params{},
		skippedNodes: &[]skippedNode{},
		index:        -1,
	}
	ctx.writermem.reset(rec)
	ctx.Writer = &ctx.writermem

	server := &handlerServer{
		trees: methodTrees{
			{method: http.MethodGet, root: getRoot},
			{method: http.MethodPost, root: postRoot},
		},
		handle: &handler{
			childMap: map[string]returnObjChild{
				"": {
					handleMethodNotAllowed: true,
					noMethod: func(ctx *Context) {
						ctx.Writer.WriteHeader(http.StatusMethodNotAllowed)
					},
					noRoute: func(ctx *Context) {
						ctx.Writer.WriteHeader(http.StatusNotFound)
					},
				},
			},
			publicGroupMiddlewares: map[string][]HandleFunc{},
		},
	}

	server.handleHTTPRequest(ctx)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status code: got %d want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("Allow header: got %q want %q", allow, http.MethodPost)
	}
	if len(*ctx.Params) != 0 {
		t.Fatalf("method probing should not leave params on context, got %v", *ctx.Params)
	}
}

func TestRedirectPreservesRawQuery(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:      httptest.NewRequest(http.MethodGet, "/users/?page=1&sort=name", nil),
		log:          nopLogger{},
		Params:       &Params{},
		skippedNodes: &[]skippedNode{},
		index:        -1,
	}
	ctx.writermem.reset(rec)
	ctx.Writer = &ctx.writermem

	server := &handlerServer{
		handle: &handler{
			publicGroupMiddlewares: map[string][]HandleFunc{},
		},
	}
	server.redirect(ctx)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status code: got %d want %d", rec.Code, http.StatusMovedPermanently)
	}
	if loc := rec.Header().Get("Location"); loc != "/users?page=1&sort=name" {
		t.Fatalf("Location header: got %q want %q", loc, "/users?page=1&sort=name")
	}
}

type bodyMediaTypeRegressionRouter struct{}

type bodyMediaTypeRegressionReq struct {
	Name string `json:"name" xml:"name"`
}

func (*bodyMediaTypeRegressionRouter) Create(input struct {
	router Router                     `paths:"/body" methods:"POST"`
	Body   bodyMediaTypeRegressionReq `body:"json"`
}) map[string]string {
	return map[string]string{"name": input.Body.Name}
}

func TestBodyMediaTypeMustMatchDeclaredContentType(t *testing.T) {
	api := New(false)
	api.SetLogger(nil)
	api.IncludeRouter(&bodyMediaTypeRegressionRouter{}, "", true)
	handler := api.Handler()

	t.Run("allows declared JSON with parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/body", strings.NewReader(`{"name":"alice"}`))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status code: got %d want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("rejects undeclared XML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/body", strings.NewReader(`<bodyMediaTypeRegressionReq><name>alice</name></bodyMediaTypeRegressionReq>`))
		req.Header.Set("Content-Type", "application/xml")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("status code: got %d want %d, body=%s", rec.Code, http.StatusUnsupportedMediaType, rec.Body.String())
		}
	})
}
