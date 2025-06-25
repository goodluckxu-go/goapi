package goapi

import (
	"net/http"
)

type Mux interface {
	AddRouter(method, path string, handler func(ctx *Context)) error
	NodFind(handler func(ctx *Context))
	MethodNotAllowed(handler func(ctx *Context))
	Static(path string, handler func(ctx *Context)) error
	http.Handler
}
