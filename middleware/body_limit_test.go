package middleware

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goodluckxu-go/goapi/v2"
)

func TestBodyLimitMiddlewareAllowsSmallBody(t *testing.T) {
	ctx, writer := newBodyLimitTestContext(http.MethodPost, "/", "small")

	BodyLimitMiddleware(10)(ctx)

	if writer.Status() != http.StatusOK {
		t.Fatalf("status: got %d want %d", writer.Status(), http.StatusOK)
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		t.Fatalf("body should be readable: %v", err)
	}
	if string(body) != "small" {
		t.Fatalf("body: got %q want %q", string(body), "small")
	}
}

func TestBodyLimitMiddlewareRejectsLargeContentLength(t *testing.T) {
	ctx, writer := newBodyLimitTestContext(http.MethodPost, "/", "too large")

	BodyLimitMiddleware(3)(ctx)

	if writer.Status() != http.StatusRequestEntityTooLarge {
		t.Fatalf("status: got %d want %d", writer.Status(), http.StatusRequestEntityTooLarge)
	}
	if !strings.Contains(writer.body.String(), defaultBodyLimitMessage) {
		t.Fatalf("body should contain default message, got %q", writer.body.String())
	}
}

func TestBodyLimitMiddlewareWrapsUnknownLengthBody(t *testing.T) {
	ctx, _ := newBodyLimitTestContext(http.MethodPost, "/", "too large")
	ctx.Request.ContentLength = -1

	BodyLimitMiddleware(3)(ctx)

	_, err := io.ReadAll(ctx.Request.Body)
	var maxBytesErr *http.MaxBytesError
	if !errors.As(err, &maxBytesErr) {
		t.Fatalf("read error: got %v want *http.MaxBytesError", err)
	}
	if maxBytesErr.Limit != 3 {
		t.Fatalf("limit: got %d want %d", maxBytesErr.Limit, 3)
	}
}

func TestBodyLimitMiddlewareReturnsRequestEntityTooLargeForUnknownLengthBody(t *testing.T) {
	api := goapi.New(false)
	api.SetLogger(nil)
	api.AddMiddleware(BodyLimitMiddleware(3))
	api.IncludeRouter(&bodyLimitAPIRouter{}, "", false)
	handler := api.Handler()

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"name":"long"}`))
	req.Header.Set("Content-Type", string(goapi.JSON))
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestBodyLimitMiddlewareAllowsEmptyBodyWithZeroLimit(t *testing.T) {
	ctx, writer := newBodyLimitTestContext(http.MethodPost, "/", "")

	BodyLimitMiddleware(0)(ctx)

	if writer.Status() != http.StatusOK {
		t.Fatalf("status: got %d want %d", writer.Status(), http.StatusOK)
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		t.Fatalf("empty body should be readable: %v", err)
	}
	if len(body) != 0 {
		t.Fatalf("body length: got %d want 0", len(body))
	}
}

func TestBodyLimitMiddlewareRejectsNegativeLimit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("negative limit should panic")
		}
	}()

	_ = BodyLimitMiddleware(-1)
}

func newBodyLimitTestContext(method, target, body string) (*goapi.Context, *bodyLimitTestWriter) {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	writer := newBodyLimitTestWriter()
	return &goapi.Context{
		Request: req,
		Writer:  writer,
	}, writer
}

type bodyLimitAPIRouter struct{}

type bodyLimitPayload struct {
	Name string `json:"name"`
}

func (*bodyLimitAPIRouter) Echo(input struct {
	Router goapi.Router     `paths:"/echo" methods:"post"`
	Body   bodyLimitPayload `body:"json"`
}) {
}

type bodyLimitTestWriter struct {
	header  http.Header
	status  int
	size    int
	written bool
	body    strings.Builder
}

func newBodyLimitTestWriter() *bodyLimitTestWriter {
	return &bodyLimitTestWriter{
		header: http.Header{},
		status: http.StatusOK,
	}
}

func (w *bodyLimitTestWriter) Header() http.Header {
	return w.header
}

func (w *bodyLimitTestWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	n, err := w.body.Write(b)
	w.size += n
	return n, err
}

func (w *bodyLimitTestWriter) WriteHeader(statusCode int) {
	if w.written {
		return
	}
	w.status = statusCode
	w.written = true
}

func (w *bodyLimitTestWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, http.ErrNotSupported
}

func (w *bodyLimitTestWriter) Flush() {}

func (w *bodyLimitTestWriter) Status() int {
	return w.status
}

func (w *bodyLimitTestWriter) Size() int {
	return w.size
}
