package goapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// nopLogger is a silent Logger implementation for tests.
type nopLogger struct{}

func (nopLogger) Debug(format string, a ...any)   {}
func (nopLogger) Info(format string, a ...any)    {}
func (nopLogger) Warning(format string, a ...any) {}
func (nopLogger) Error(format string, a ...any)   {}
func (nopLogger) Fatal(format string, a ...any)   {}

// newTestContext builds a Context wired like production (recorder + request).
func newTestContext(t *testing.T, req *http.Request) *Context {
	t.Helper()
	rec := httptest.NewRecorder()
	ctx := &Context{
		Request: req,
		log:     nopLogger{},
	}
	ctx.writermem.ResponseWriter = rec
	ctx.Writer = &ctx.writermem
	return ctx
}

func TestContext_SetGet(t *testing.T) {
	ctx := newTestContext(t, httptest.NewRequest(http.MethodGet, "/", nil))

	ctx.Set("k1", "v1")
	v, ok := ctx.Get("k1")
	if !ok || v != "v1" {
		t.Fatalf("Get: want (v1, true), got (%v, %v)", v, ok)
	}
	_, ok = ctx.Get("missing")
	if ok {
		t.Fatal("Get: expected false for missing key")
	}
}

func TestContext_Value(t *testing.T) {
	t.Run("string key from Values map", func(t *testing.T) {
		ctx := newTestContext(t, httptest.NewRequest(http.MethodGet, "/", nil))
		ctx.Set("trace", "abc")
		if got := ctx.Value("trace"); got != "abc" {
			t.Fatalf("Value(trace): want abc, got %v", got)
		}
	})

	t.Run("string key falls back to request context", func(t *testing.T) {
		type ctxKey struct{}
		k := ctxKey{}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), k, "from-req"))
		ctx := newTestContext(t, req)
		if got := ctx.Value(k); got != "from-req" {
			t.Fatalf("Value(custom key): want from-req, got %v", got)
		}
	})

	t.Run("nil request returns nil for unknown string key", func(t *testing.T) {
		ctx := &Context{}
		if got := ctx.Value("x"); got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})
}

func TestContext_FullPath(t *testing.T) {
	ctx := &Context{}
	ctx.fullPath = "/api/v1/users"
	if ctx.FullPath() != "/api/v1/users" {
		t.Fatalf("FullPath: want /api/v1/users, got %q", ctx.FullPath())
	}
}

func TestContext_reset(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/p?q=1", nil)
	ctx := newTestContext(t, req)
	ctx.Request = req
	ctx.Params = &Params{{Key: "id", Value: "1"}}
	ctx.skippedNodes = skippedNodes
	ctx.Set("x", 1)
	ctx.handlers = []HandleFunc{fakeHandler}
	ctx.index = 3
	ctx.fullPath = "/old"
	ctx.queryCache = url.Values{"q": {"1"}}
	ctx.ChildPath = "child"
	ctx.RequestID = "rid"

	ctx.reset()

	if ctx.Writer != &ctx.writermem {
		t.Fatal("reset: Writer should point to writermem")
	}
	if len(*ctx.Params) != 0 {
		t.Fatalf("reset: Params should be empty, got %v", *ctx.Params)
	}
	if _, ok := ctx.Get("x"); ok {
		t.Fatal("reset: Values should be cleared")
	}
	if len(ctx.handlers) != 0 {
		t.Fatal("reset: handlers slice should be cleared")
	}
	if ctx.index != -1 {
		t.Fatalf("reset: index want -1, got %d", ctx.index)
	}
	if ctx.fullPath != "" {
		t.Fatalf("reset: fullPath want empty, got %q", ctx.fullPath)
	}
	if ctx.queryCache != nil {
		t.Fatal("reset: queryCache should be nil")
	}
	if ctx.ChildPath != "" || ctx.RequestID != "" {
		t.Fatalf("reset: ChildPath/RequestID should be empty, got %q / %q", ctx.ChildPath, ctx.RequestID)
	}
}

func TestContext_Next(t *testing.T) {
	t.Run("runs handlers in order", func(t *testing.T) {
		ctx := newTestContext(t, httptest.NewRequest(http.MethodGet, "/", nil))
		var order []int
		ctx.handleExcept = func(*Context, string, ...int) { t.Fatal("handleExcept should not run") }
		ctx.handlers = []HandleFunc{
			func(c *Context) {
				order = append(order, 1)
				c.Next()
			},
			func(c *Context) { order = append(order, 2) },
		}
		ctx.index = -1
		ctx.Next()
		if len(order) != 2 || order[0] != 1 || order[1] != 2 {
			t.Fatalf("handler order: want [1 2], got %v", order)
		}
	})

	t.Run("nil handler is skipped", func(t *testing.T) {
		ctx := newTestContext(t, httptest.NewRequest(http.MethodGet, "/", nil))
		var ran bool
		ctx.handleExcept = func(*Context, string, ...int) { t.Fatal("handleExcept should not run") }
		ctx.handlers = []HandleFunc{
			nil,
			func(c *Context) { ran = true },
		}
		ctx.index = -1
		ctx.Next()
		if !ran {
			t.Fatal("second handler should run after nil is skipped")
		}
	})

	t.Run("panic invokes handleExcept", func(t *testing.T) {
		ctx := newTestContext(t, httptest.NewRequest(http.MethodGet, "/", nil))
		var got string
		ctx.handleExcept = func(_ *Context, err string, _ ...int) { got = err }
		ctx.handlers = []HandleFunc{
			func(c *Context) { panic("boom") },
		}
		ctx.index = -1
		ctx.Next()
		if got != "boom" {
			t.Fatalf("handleExcept err: want boom, got %q", got)
		}
	})
}

func TestContext_Logger(t *testing.T) {
	ctx := &Context{log: nopLogger{}}
	if ctx.Logger() == nil {
		t.Fatal("Logger() should not be nil")
	}
}

func TestContext_RemoteIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "198.51.100.7:12345"
	ctx := newTestContext(t, req)
	if ip := ctx.RemoteIP(); ip != "198.51.100.7" {
		t.Fatalf("RemoteIP: want 198.51.100.7, got %q", ip)
	}
}

func TestContext_ClientIP(t *testing.T) {
	t.Run("uses X-Forwarded-For when valid IPv4", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:8080"
		req.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1")
		ctx := newTestContext(t, req)
		if got := ctx.ClientIP(); got != "203.0.113.9" {
			t.Fatalf("ClientIP: want 203.0.113.9, got %q", got)
		}
	})

	t.Run("falls back to X-Real-IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:8080"
		req.Header.Set("X-Real-IP", "198.51.100.1")
		ctx := newTestContext(t, req)
		if got := ctx.ClientIP(); got != "198.51.100.1" {
			t.Fatalf("ClientIP: want 198.51.100.1, got %q", got)
		}
	})

	t.Run("falls back to RemoteAddr when headers missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.1:443"
		ctx := newTestContext(t, req)
		if got := ctx.ClientIP(); got != "192.0.2.1" {
			t.Fatalf("ClientIP: want 192.0.2.1, got %q", got)
		}
	})
}

func TestContext_Query(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search?q=hello&tag=go", nil)
	ctx := newTestContext(t, req)
	q := ctx.Query()
	if q.Get("q") != "hello" || q.Get("tag") != "go" {
		t.Fatalf("Query: unexpected values: %v", q)
	}
	// url.Values is a map type and cannot be compared with ==; mutating the first
	// result must be visible on the second call if the cache is the same instance.
	q.Set("cached", "yes")
	if ctx.Query().Get("cached") != "yes" {
		t.Fatal("Query: expected second call to return the same cached url.Values")
	}
}

func TestContext_RequestContext(t *testing.T) {
	t.Run("Deadline Done Err with background context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := newTestContext(t, req)
		if _, ok := ctx.Deadline(); ok {
			t.Fatal("Deadline: expected ok false for background context")
		}
		// context.Background().Done() is nil in the standard library: there is no
		// cancelation channel for a non-cancelable context. Context.Done must match
		// the underlying request context, not assume a non-nil channel.
		if ctx.Done() != req.Context().Done() {
			t.Fatal("Done: should delegate to request.Context().Done()")
		}
		if ctx.Err() != nil {
			t.Fatalf("Err: want nil, got %v", ctx.Err())
		}
	})

	t.Run("Done delegates for cancelable request context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c, cancel := context.WithCancel(req.Context())
		defer cancel()
		req = req.WithContext(c)
		ctx := newTestContext(t, req)
		if ctx.Done() == nil {
			t.Fatal("Done: WithCancel should yield a non-nil channel")
		}
	})

	t.Run("Deadline with timeout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c, cancel := context.WithTimeout(req.Context(), time.Hour)
		defer cancel()
		req = req.WithContext(c)
		ctx := newTestContext(t, req)
		_, ok := ctx.Deadline()
		if !ok {
			t.Fatal("Deadline: expected ok true when request has timeout")
		}
	})
}
