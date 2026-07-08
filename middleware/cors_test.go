package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goodluckxu-go/goapi/v2"
)

type corsTestRouter struct{}

func (*corsTestRouter) Ping(input struct {
	Router goapi.Router `paths:"/ping" methods:"get"`
}) {
}

func TestCORSMiddlewareAllowsSimpleRequest(t *testing.T) {
	handler := newCORSTestHandler(t, CORSMiddleware())

	rec := runCORSRequest(handler, http.MethodGet, "https://example.com", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(headerAccessControlAllowOrigin); got != "*" {
		t.Fatalf("allow origin: got %q want %q", got, "*")
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	handler := newCORSTestHandler(t, CORSMiddlewareWithConfig(CORSConfig{
		MaxAge: 10 * time.Minute,
	}))

	rec := runCORSRequest(handler, http.MethodOptions, "https://example.com", http.MethodPut, "X-Trace-ID")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get(headerAccessControlAllowOrigin); got != "*" {
		t.Fatalf("allow origin: got %q want %q", got, "*")
	}
	if got := rec.Header().Get(headerAccessControlAllowHeaders); got != "X-Trace-ID" {
		t.Fatalf("allow headers: got %q want %q", got, "X-Trace-ID")
	}
	if got := rec.Header().Get(headerAccessControlMaxAge); got != "600" {
		t.Fatalf("max age: got %q want %q", got, "600")
	}
}

func TestCORSMiddlewareRejectsDisallowedPreflightMethod(t *testing.T) {
	handler := newCORSTestHandler(t, CORSMiddlewareWithConfig(CORSConfig{
		AllowMethods: []string{http.MethodGet},
	}))

	rec := runCORSRequest(handler, http.MethodOptions, "https://example.com", http.MethodPut, "X-Trace-ID")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusForbidden)
	}
	if got := rec.Header().Get(headerAccessControlAllowOrigin); got != "" {
		t.Fatalf("allow origin should be empty, got %q", got)
	}
}

func TestCORSMiddlewareRejectsDisallowedPreflightHeader(t *testing.T) {
	handler := newCORSTestHandler(t, CORSMiddlewareWithConfig(CORSConfig{
		AllowMethods: []string{http.MethodPut},
		AllowHeaders: []string{"Content-Type"},
	}))

	rec := runCORSRequest(handler, http.MethodOptions, "https://example.com", http.MethodPut, "X-Trace-ID")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusForbidden)
	}
	if got := rec.Header().Get(headerAccessControlAllowOrigin); got != "" {
		t.Fatalf("allow origin should be empty, got %q", got)
	}
}

func TestCORSMiddlewareAllowsCredentialsForSpecificOrigin(t *testing.T) {
	handler := newCORSTestHandler(t, CORSMiddlewareWithConfig(CORSConfig{
		AllowOrigins:     []string{"https://app.example.com"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Request-ID"},
	}))

	rec := runCORSRequest(handler, http.MethodGet, "https://app.example.com", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(headerAccessControlAllowOrigin); got != "https://app.example.com" {
		t.Fatalf("allow origin: got %q want %q", got, "https://app.example.com")
	}
	if got := rec.Header().Get(headerAccessControlAllowCredentials); got != "true" {
		t.Fatalf("allow credentials: got %q want %q", got, "true")
	}
	if got := rec.Header().Get(headerAccessControlExposeHeaders); got != "X-Request-ID" {
		t.Fatalf("expose headers: got %q want %q", got, "X-Request-ID")
	}
}

func TestCORSMiddlewareCredentialsRequireExplicitOrigins(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("AllowCredentials without explicit origin should panic")
		}
	}()

	_ = CORSMiddlewareWithConfig(CORSConfig{
		AllowCredentials: true,
	})
}

func TestCORSMiddlewareCredentialsRejectWildcardOrigin(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("AllowCredentials with wildcard origin should panic")
		}
	}()

	_ = CORSMiddlewareWithConfig(CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	})
}

func TestCORSMiddlewareRejectsDisallowedPreflight(t *testing.T) {
	handler := newCORSTestHandler(t, CORSMiddlewareWithConfig(CORSConfig{
		AllowOrigins: []string{"https://app.example.com"},
	}))

	rec := runCORSRequest(handler, http.MethodOptions, "https://evil.example.com", http.MethodPut, "X-Trace-ID")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusForbidden)
	}
	if got := rec.Header().Get(headerAccessControlAllowOrigin); got != "" {
		t.Fatalf("allow origin should be empty, got %q", got)
	}
}

func newCORSTestHandler(t *testing.T, middleware goapi.HandleFunc) http.Handler {
	t.Helper()

	api := goapi.New(false)
	api.SetLogger(nil)
	api.AddMiddleware(middleware)
	api.IncludeRouter(&corsTestRouter{}, "", false)
	return api.Handler()
}

func runCORSRequest(handler http.Handler, method, origin, requestMethod, requestHeaders string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/ping", nil)
	req.RemoteAddr = "198.51.100.10:1234"
	if origin != "" {
		req.Header.Set(headerOrigin, origin)
	}
	if requestMethod != "" {
		req.Header.Set(headerAccessControlRequestMethod, requestMethod)
	}
	if requestHeaders != "" {
		req.Header.Set(headerAccessControlRequestHeaders, requestHeaders)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
