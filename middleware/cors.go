package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goodluckxu-go/goapi/v2"
)

const (
	headerOrigin                        = "Origin"
	headerVary                          = "Vary"
	headerAccessControlRequestMethod    = "Access-Control-Request-Method"
	headerAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	headerAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	headerAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	headerAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	headerAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	headerAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	headerAccessControlMaxAge           = "Access-Control-Max-Age"
)

// CORSConfig configures the CORS middleware.
type CORSConfig struct {
	// AllowOrigins is the list of origins allowed to access the resource.
	// If empty and AllowOriginFunc is nil, all origins are allowed.
	// AllowCredentials requires explicit origins or AllowOriginFunc.
	AllowOrigins []string
	// AllowMethods is the list of methods allowed for preflight requests.
	// If empty, common HTTP methods are allowed.
	AllowMethods []string
	// AllowHeaders is the list of headers allowed for preflight requests.
	// If empty or set to "*", requested headers are echoed.
	AllowHeaders []string
	// ExposeHeaders is the list of response headers exposed to browsers.
	ExposeHeaders []string
	// AllowCredentials sets Access-Control-Allow-Credentials to true.
	AllowCredentials bool
	// MaxAge is the preflight cache duration.
	MaxAge time.Duration
	// AllowOriginFunc can be used to allow origins dynamically.
	AllowOriginFunc func(origin string) bool
}

// CORSMiddleware returns a permissive CORS middleware.
func CORSMiddleware() goapi.HandleFunc {
	return CORSMiddlewareWithConfig(CORSConfig{})
}

// CORSMiddlewareWithConfig returns a configurable CORS middleware.
func CORSMiddlewareWithConfig(config CORSConfig) goapi.HandleFunc {
	allowOrigins := normalizeHeaderValues(config.AllowOrigins)
	if config.AllowCredentials && len(allowOrigins) == 0 && config.AllowOriginFunc == nil {
		panic("goapi: CORS AllowCredentials requires AllowOrigins or AllowOriginFunc")
	}
	if config.AllowCredentials && containsHeaderValue(allowOrigins, "*") {
		panic("goapi: CORS AllowCredentials cannot be used with wildcard AllowOrigins")
	}
	if len(allowOrigins) == 0 && config.AllowOriginFunc == nil {
		allowOrigins = []string{"*"}
	}
	allowMethods := normalizeHeaderValues(config.AllowMethods)
	if len(allowMethods) == 0 {
		allowMethods = []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		}
	}
	allowHeaders := normalizeHeaderValues(config.AllowHeaders)
	exposeHeaders := normalizeHeaderValues(config.ExposeHeaders)
	allowAllOrigins := containsHeaderValue(allowOrigins, "*")
	allowOriginSet := make(map[string]struct{}, len(allowOrigins))
	for _, origin := range allowOrigins {
		allowOriginSet[origin] = struct{}{}
	}

	return func(ctx *goapi.Context) {
		origin := ctx.Request.Header.Get(headerOrigin)
		if origin == "" {
			ctx.Next()
			return
		}

		isPreflight := ctx.Request.Method == http.MethodOptions &&
			strings.TrimSpace(ctx.Request.Header.Get(headerAccessControlRequestMethod)) != ""
		allowOrigin := resolveAllowOrigin(origin, allowAllOrigins, allowOriginSet, config.AllowOriginFunc, config.AllowCredentials)
		if allowOrigin == "" {
			if isPreflight {
				http.Error(ctx.Writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			ctx.Next()
			return
		}

		if isPreflight && !isPreflightAllowed(ctx, allowMethods, allowHeaders) {
			http.Error(ctx.Writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		headers := ctx.Writer.Header()
		headers.Set(headerAccessControlAllowOrigin, allowOrigin)
		addVary(headers, headerOrigin)
		if config.AllowCredentials {
			headers.Set(headerAccessControlAllowCredentials, "true")
		}
		if len(exposeHeaders) > 0 {
			headers.Set(headerAccessControlExposeHeaders, strings.Join(exposeHeaders, ", "))
		}

		if !isPreflight {
			ctx.Next()
			return
		}

		addVary(headers, headerAccessControlRequestMethod)
		addVary(headers, headerAccessControlRequestHeaders)
		headers.Set(headerAccessControlAllowMethods, strings.Join(allowMethods, ", "))
		setAllowHeaders(headers, allowHeaders, ctx.Request.Header.Get(headerAccessControlRequestHeaders))
		if config.MaxAge > 0 {
			headers.Set(headerAccessControlMaxAge, strconv.Itoa(int(config.MaxAge.Seconds())))
		}
		ctx.Writer.WriteHeader(http.StatusNoContent)
	}
}

func isPreflightAllowed(ctx *goapi.Context, allowMethods, allowHeaders []string) bool {
	requestMethod := strings.TrimSpace(ctx.Request.Header.Get(headerAccessControlRequestMethod))
	if !containsHeaderValueFold(allowMethods, requestMethod) {
		return false
	}

	if len(allowHeaders) == 0 || containsHeaderValue(allowHeaders, "*") {
		return true
	}
	for _, header := range splitHeaderList(ctx.Request.Header.Get(headerAccessControlRequestHeaders)) {
		if !containsHeaderValueFold(allowHeaders, header) {
			return false
		}
	}
	return true
}

func resolveAllowOrigin(origin string, allowAll bool, allowOriginSet map[string]struct{}, allowOriginFunc func(string) bool, allowCredentials bool) string {
	if allowOriginFunc != nil && allowOriginFunc(origin) {
		return origin
	}
	if allowAll {
		if allowCredentials {
			return origin
		}
		return "*"
	}
	if _, ok := allowOriginSet[origin]; ok {
		return origin
	}
	return ""
}

func setAllowHeaders(headers http.Header, allowHeaders []string, requestedHeaders string) {
	if len(allowHeaders) == 0 || containsHeaderValue(allowHeaders, "*") {
		if requestedHeaders != "" {
			headers.Set(headerAccessControlAllowHeaders, requestedHeaders)
		} else {
			headers.Set(headerAccessControlAllowHeaders, "*")
		}
		return
	}
	headers.Set(headerAccessControlAllowHeaders, strings.Join(allowHeaders, ", "))
}

func normalizeHeaderValues(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func containsHeaderValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsHeaderValueFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func splitHeaderList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func addVary(headers http.Header, value string) {
	for _, existing := range headers.Values(headerVary) {
		for _, item := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(item), value) {
				return
			}
		}
	}
	headers.Add(headerVary, value)
}
