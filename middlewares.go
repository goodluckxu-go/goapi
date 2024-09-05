package goapi

import (
	"fmt"
	"net/http"
	"time"
)

func notFind() func(ctx *Context) {
	return func(ctx *Context) {
		http.NotFound(ctx.Writer, ctx.Request)
		if lg, ok := ctx.log.(*levelHandleLogger); ok && lg.log == nil {
			return
		}
		status := "404"
		statusText := http.StatusText(404)
		if IsDefaultLogger(ctx.log) {
			status = colorError(status)
			statusText = colorError(statusText)
		}
		ctx.Logger().Info("[0.000ms] %v - \"%v %v\" %v %v", ctx.Request.RemoteAddr,
			ctx.Request.Method, ctx.Request.URL.Path, status, statusText)
	}
}

func setLogger() func(ctx *Context) {
	return func(ctx *Context) {
		if lg, ok := ctx.log.(*levelHandleLogger); ok && lg.log == nil {
			ctx.Next()
			return
		}
		begin := time.Now()
		ctx.Next()
		elapsed := time.Since(begin)
		if resp, ok := ctx.Writer.(*ResponseWriter); ok {
			status := fmt.Sprintf("%v", resp.Status())
			statusText := http.StatusText(resp.Status())
			if IsDefaultLogger(ctx.log) && len(status) == 3 {
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
			ctx.Logger().Info("[%.3fms] %v - \"%v %v\" %v %v", float64(elapsed.Nanoseconds())/1e6, ctx.Request.RemoteAddr,
				ctx.Request.Method, ctx.Request.URL.Path, status, statusText)
		}
	}
}
