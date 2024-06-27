package goapi

import (
	"fmt"
	"net/http"
	"time"
)

func notFind() func(ctx *Context) {
	return func(ctx *Context) {
		http.NotFound(ctx.Writer, ctx.Request)
		ctx.Log.Info("[0.000ms] %v - \"%v %v\" %v %v", ctx.Request.RemoteAddr,
			ctx.Request.Method, ctx.Request.URL.Path, colorError("404"), colorError(http.StatusText(404)))
	}
}

func setLogger() func(ctx *Context) {
	return func(ctx *Context) {
		begin := time.Now()
		ctx.Next()
		elapsed := time.Since(begin)
		if resp, ok := ctx.Writer.(*ResponseWriter); ok {
			status := fmt.Sprintf("%v", resp.Status())
			statusText := http.StatusText(resp.Status())
			if len(status) == 3 {
				if status[0] == '1' || status[0] == '2' {
					status = colorInfo(status)
					statusText = colorInfo(statusText)
				} else if status[0] == '4' || status[0] == '5' {
					status = colorError(status)
					statusText = colorError(statusText)
				} else if status[0] == '3' {
					status = colorWarning(status)
					statusText = colorWarning(statusText)
				}
			}
			ctx.Log.Info("[%.3fms] %v - \"%v %v\" %v %v", float64(elapsed.Nanoseconds())/1e6, ctx.Request.RemoteAddr,
				ctx.Request.Method, ctx.Request.URL.Path, status, statusText)
		}
	}
}
