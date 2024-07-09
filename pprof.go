package goapi

import "net/http/pprof"

type pprofInfo struct {
}

func (p *pprofInfo) Pprof(ctx *Context, input struct {
	router Router `path:"/pprof/{path}" method:"get" tags:"pprof"`
	Path   string `path:"path"`
}) {
	pprof.Index(ctx.Writer, ctx.Request)
}

func (p *pprofInfo) PprofIndex(ctx *Context, input struct {
	router Router `path:"/pprof/" method:"get" tags:"pprof"`
}) {
	pprof.Index(ctx.Writer, ctx.Request)
}

func (p *pprofInfo) PprofCmdline(ctx *Context, input struct {
	router Router `path:"/pprof/cmdline" method:"get" tags:"pprof"`
}) {
	pprof.Cmdline(ctx.Writer, ctx.Request)
}

func (p *pprofInfo) PprofProfile(ctx *Context, input struct {
	router Router `path:"/pprof/profile" method:"get" tags:"pprof"`
}) {
	pprof.Profile(ctx.Writer, ctx.Request)
}

func (p *pprofInfo) PprofSymbol(ctx *Context, input struct {
	router Router `path:"/pprof/symbol" method:"get" tags:"pprof"`
}) {
	pprof.Symbol(ctx.Writer, ctx.Request)
}

func (p *pprofInfo) PprofTrace(ctx *Context, input struct {
	router Router `path:"/pprof/trace" method:"get" tags:"pprof"`
}) {
	pprof.Trace(ctx.Writer, ctx.Request)
}
