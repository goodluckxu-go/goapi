package goapi

import (
	"net/http/pprof"
)

func debugPprof(ctx *Context, input struct {
	router Router `paths:"/pprof/,/pprof/{path}" methods:"get" tags:"pprof"`
	Path   string `path:"path"`
}) {
	switch input.Path {
	case "":
		pprof.Index(ctx.Writer, ctx.Request)
	case "cmdline":
		pprof.Cmdline(ctx.Writer, ctx.Request)
	case "profile":
		pprof.Profile(ctx.Writer, ctx.Request)
	case "symbol":
		pprof.Symbol(ctx.Writer, ctx.Request)
	case "trace":
		pprof.Trace(ctx.Writer, ctx.Request)
	default:
		pprof.Index(ctx.Writer, ctx.Request)
	}
}
