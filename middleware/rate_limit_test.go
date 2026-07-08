package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goodluckxu-go/goapi/v2"
)

type rateLimitTestRouter struct{}

func (*rateLimitTestRouter) Ping(input struct {
	Router goapi.Router `paths:"/ping" methods:"get"`
}) {
}

func TestRateLimitMiddlewareBlocksSameKey(t *testing.T) {
	handler := newRateLimitTestHandler(t, RateLimitMiddleware(1, time.Hour))

	first := runRateLimitRequest(handler, "198.51.100.1:1234", "", "")
	if first.Code != http.StatusOK {
		t.Fatalf("first request status: got %d want %d", first.Code, http.StatusOK)
	}

	second := runRateLimitRequest(handler, "198.51.100.1:1234", "", "")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status: got %d want %d", second.Code, http.StatusTooManyRequests)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header should be set")
	}
}

func TestRateLimitMiddlewareIsolatesKeys(t *testing.T) {
	handler := newRateLimitTestHandler(t, RateLimitMiddleware(1, time.Hour))

	first := runRateLimitRequest(handler, "198.51.100.1:1234", "", "")
	if first.Code != http.StatusOK {
		t.Fatalf("first request status: got %d want %d", first.Code, http.StatusOK)
	}

	second := runRateLimitRequest(handler, "198.51.100.2:1234", "", "")
	if second.Code != http.StatusOK {
		t.Fatalf("second key status: got %d want %d", second.Code, http.StatusOK)
	}
}

func TestRateLimitMiddlewareDefaultKeyIgnoresForwardedHeaders(t *testing.T) {
	handler := newRateLimitTestHandler(t, RateLimitMiddleware(1, time.Hour))

	first := runRateLimitRequest(handler, "198.51.100.1:1234", "X-Forwarded-For", "203.0.113.1")
	if first.Code != http.StatusOK {
		t.Fatalf("first request status: got %d want %d", first.Code, http.StatusOK)
	}

	second := runRateLimitRequest(handler, "198.51.100.1:1234", "X-Forwarded-For", "203.0.113.2")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("spoofed forwarded header status: got %d want %d", second.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiterRefillsOverTime(t *testing.T) {
	limiter := newRateLimiter(RateLimitConfig{
		Limit:  1,
		Window: time.Second,
	})
	now := time.Unix(100, 0)

	if ok, _ := limiter.allow("client", now); !ok {
		t.Fatal("first request should be allowed")
	}
	if ok, _ := limiter.allow("client", now); ok {
		t.Fatal("second request should be blocked")
	}
	if ok, _ := limiter.allow("client", now.Add(time.Second)); !ok {
		t.Fatal("request after refill window should be allowed")
	}
}

func TestRateLimiterMaxKeysRejectsNewKeys(t *testing.T) {
	limiter := newRateLimiter(RateLimitConfig{
		Limit:   2,
		Window:  time.Hour,
		MaxKeys: 1,
	})
	now := time.Unix(100, 0)

	if ok, _ := limiter.allow("client-1", now); !ok {
		t.Fatal("first key should be allowed")
	}
	if ok, _ := limiter.allow("client-2", now); ok {
		t.Fatal("new key should be rejected after MaxKeys is reached")
	}
	if ok, _ := limiter.allow("client-1", now); !ok {
		t.Fatal("existing key should still be allowed")
	}
}

func TestRateLimitMiddlewareUsesCustomKeyFunc(t *testing.T) {
	handler := newRateLimitTestHandler(t, RateLimitMiddlewareWithConfig(RateLimitConfig{
		Limit:  1,
		Window: time.Hour,
		KeyFunc: func(ctx *goapi.Context) string {
			return ctx.Request.Header.Get("X-User-ID")
		},
	}))

	first := runRateLimitRequest(handler, "198.51.100.1:1234", "X-User-ID", "u1")
	if first.Code != http.StatusOK {
		t.Fatalf("first request status: got %d want %d", first.Code, http.StatusOK)
	}

	second := runRateLimitRequest(handler, "198.51.100.2:1234", "X-User-ID", "u1")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("same custom key status: got %d want %d", second.Code, http.StatusTooManyRequests)
	}
}

func newRateLimitTestHandler(t *testing.T, middleware goapi.HandleFunc) http.Handler {
	t.Helper()

	api := goapi.New(false)
	api.SetLogger(nil)
	api.AddMiddleware(middleware)
	api.IncludeRouter(&rateLimitTestRouter{}, "", false)
	return api.Handler()
}

func runRateLimitRequest(handler http.Handler, remoteAddr, headerKey, headerValue string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = remoteAddr
	if headerKey != "" {
		req.Header.Set(headerKey, headerValue)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
