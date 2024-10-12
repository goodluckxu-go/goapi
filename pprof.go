package goapi

import "net/http/pprof"

type pprofInfo struct {
}

func (p *pprofInfo) PprofIndex(ctx *Context, input struct {
	router Router `path:"/pprof/" method:"get" tags:"pprof"`
}) {
	pprof.Index(ctx.Writer, ctx.Request)
}

func (p *pprofInfo) Pprof(ctx *Context, input struct {
	router Router `path:"/pprof/{path}" method:"get" tags:"pprof"`
	Path   string `path:"path"`
}) {
	switch input.Path {
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
