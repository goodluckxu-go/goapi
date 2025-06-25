package goapi

import (
	"net/http"
	"sync"
)

func newGoAPIMux() *goAPIMux {
	mux := &goAPIMux{
		routers: map[string]*node{},
	}
	mux.pool.New = func() any {
		return &Context{}
	}
	return mux
}

type goAPIMux struct {
	routers map[string]*node
	log     Logger
	notFind *appRouter
	pool    sync.Pool
}

// ServeHTTP Implement http.Handler interface
func (m *goAPIMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := m.pool.Get().(*Context)
	ctx.writermem.reset(w)
	ctx.reset()
	ctx.Request = r
	m.handleHTTPRequest(ctx)
	m.pool.Put(ctx)
}

func (m *goAPIMux) AddRouter(method, path string, handler func(ctx *Context)) (err error) {
	tree := m.routers[method]
	if tree == nil {
		tree = &node{}
	}
	if err = tree.addRouter(path, &appRouter{
		path:    path,
		method:  method,
		handler: handler,
	}); err != nil {
		return
	}
	m.routers[method] = tree
	return
}

func (m *goAPIMux) NodFind(handler func(ctx *Context)) {
	m.notFind = &appRouter{
		handler: handler,
	}
}

func (m *goAPIMux) MethodNotAllowed(handler func(ctx *Context)) {

}

func (m *goAPIMux) Static(path string, handler func(ctx *Context)) (err error) {
	tree := m.routers[http.MethodGet]
	if tree == nil {
		tree = &node{}
	}
	if err = tree.addRouter(path, &appRouter{
		path:     path,
		isPrefix: true,
		method:   http.MethodGet,
		handler:  handler,
	}); err != nil {
		return
	}
	m.routers[http.MethodGet] = tree
	return
}

func (m *goAPIMux) handleHTTPRequest(ctx *Context) {
	router, paths, exists := m.searchRouters(ctx.Request.URL.Path, ctx.Request.Method)
	if !exists {
		m.notFind.handler(ctx)
		return
	}
	ctx.paths = paths
	ctx.fullPath = router.path
	router.handler(ctx)
}

func (m *goAPIMux) searchRouters(urlPath, method string) (router *appRouter, paths map[string]string, exists bool) {
	tree := m.routers[method]
	if tree == nil {
		return
	}
	return tree.findRouter(urlPath)
}
