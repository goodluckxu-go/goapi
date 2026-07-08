package middleware

import (
	"net/http"

	"github.com/goodluckxu-go/goapi/v2"
)

const defaultBodyLimitMessage = "request body too large"

// BodyLimitConfig configures the request body size limit middleware.
type BodyLimitConfig struct {
	// Limit is the maximum number of bytes allowed in the request body.
	// A zero limit allows only empty request bodies.
	Limit int64
	// Message is written in the 413 response body.
	Message string
}

// BodyLimitMiddleware limits the request body size.
func BodyLimitMiddleware(limit int64) goapi.HandleFunc {
	return BodyLimitMiddlewareWithConfig(BodyLimitConfig{
		Limit: limit,
	})
}

// BodyLimitMiddlewareWithConfig returns a configurable request body size limit middleware.
func BodyLimitMiddlewareWithConfig(config BodyLimitConfig) goapi.HandleFunc {
	if config.Limit < 0 {
		panic("goapi: body limit Limit must be greater than or equal to zero")
	}
	message := config.Message
	if message == "" {
		message = defaultBodyLimitMessage
	}

	return func(ctx *goapi.Context) {
		if ctx == nil || ctx.Request == nil {
			return
		}
		if ctx.Request.Body == nil {
			ctx.Next()
			return
		}
		if ctx.Request.ContentLength > config.Limit {
			http.Error(ctx.Writer, message, http.StatusRequestEntityTooLarge)
			return
		}
		if ctx.Writer != nil {
			ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, config.Limit)
		}
		ctx.Next()
	}
}
