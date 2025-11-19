package goapi

import (
	"net/http"
	"time"
)

func setLogger() func(ctx *Context) {
	return func(ctx *Context) {
		if lg, ok := ctx.log.(*levelHandleLogger); ok && lg.log == nil {
			ctx.Next()
			return
		}
		begin := time.Now()
		ctx.Next()
		elapsed := time.Since(begin)
		status := toString(ctx.Writer.Status())
		statusText := http.StatusText(ctx.Writer.Status())
		if isDefaultLogger(ctx.log) && len(status) == 3 {
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
		ctx.Logger().Info("[%.3fms] %v - \"%v %v\" %v %v", float64(elapsed.Nanoseconds())/1e6, ctx.ClientIP(),
			ctx.Request.Method, ctx.Request.URL.Path, status, statusText)
	}
}
