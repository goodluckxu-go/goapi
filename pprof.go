package goapi

import (
	"net/http/pprof"
)

type pprofInfo struct {
}

func (p *pprofInfo) Pprof(ctx *Context, input struct {
	router Router `path:"/pprof/,/pprof/{path}" method:"get" tags:"pprof"`
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
