package goapi

import (
	"encoding/json"
	"errors"
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

type copyLoggerRegression struct {
	Values [2]int
	Items  []string
	Labels map[string]int
	ctx    *Context
}

func (l *copyLoggerRegression) Debug(format string, a ...any)   {}
func (l *copyLoggerRegression) Info(format string, a ...any)    {}
func (l *copyLoggerRegression) Warning(format string, a ...any) {}
func (l *copyLoggerRegression) Error(format string, a ...any)   {}
func (l *copyLoggerRegression) Fatal(format string, a ...any)   {}
func (l *copyLoggerRegression) SetContext(ctx *Context)         { l.ctx = ctx }

func TestCopyLoggerCompositeFields(t *testing.T) {
	src := &copyLoggerRegression{
		Values: [2]int{1, 2},
		Items:  []string{"a", "b"},
		Labels: map[string]int{"a": 1},
	}

	got := (&handlerServer{}).copyLogger(src).(*copyLoggerRegression)
	if got.Values != src.Values {
		t.Fatalf("array field was not copied: got %v want %v", got.Values, src.Values)
	}
	if !reflect.DeepEqual(got.Items, src.Items) {
		t.Fatalf("slice field was not copied: got %v want %v", got.Items, src.Items)
	}
	if !reflect.DeepEqual(got.Labels, src.Labels) {
		t.Fatalf("map field was not copied: got %v want %v", got.Labels, src.Labels)
	}

	got.Items[0] = "changed"
	if src.Items[0] == "changed" {
		t.Fatal("slice field should be copied into a new slice")
	}
	got.Labels["a"] = 2
	if src.Labels["a"] == 2 {
		t.Fatal("map field should be copied into a new map")
	}
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
