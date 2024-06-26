package goapi

import (
	"net/http"
	"time"
)

func notFind() func(ctx *Context) {
	return func(ctx *Context) {
		http.NotFound(ctx.Writer, ctx.Request)
		ctx.Log.Info("[0.000ms] %v - \"%v %v\" %v %v", ctx.Request.RemoteAddr, ctx.Request.Method,
			ctx.Request.URL.Path, 404, http.StatusText(404))
	}
}

func setLogger() func(ctx *Context) {
	return func(ctx *Context) {
		begin := time.Now()
		ctx.Next()
		elapsed := time.Since(begin)
		if resp, ok := ctx.Writer.(*ResponseWriter); ok {
			ctx.Log.Info("[%.3fms] %v - \"%v %v\" %v %v", float64(elapsed.Nanoseconds())/1e6, ctx.Request.RemoteAddr,
				ctx.Request.Method, ctx.Request.URL.Path, resp.Status(), http.StatusText(resp.Status()))
		}
	}
}
