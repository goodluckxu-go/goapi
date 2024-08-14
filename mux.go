package goapi

import (
	"net/http"
	"strings"
)

func newGoAPIMux(log Logger) *goAPIMux {
	return &goAPIMux{
		routers: map[string]*node{},
		statics: map[string][]*appRouter{},
		log:     log,
	}
}

type goAPIMux struct {
	routers map[string]*node
	statics map[string][]*appRouter
	log     Logger
	notFind *appRouter
}

// ServeHTTP Implement http.Handler interface
func (m *goAPIMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		index:   -1,
		log:     m.log,
		Request: r,
		Writer:  &ResponseWriter{ResponseWriter: w},
	}
	m.handleHTTPRequest(ctx)
}

func (m *goAPIMux) addRouters(path, method string, router *appRouter) (err error) {
	if router.isPrefix {
		m.statics[method] = append(m.statics[method], router)
	}
	tree := m.routers[method]
	if tree == nil {
		tree = &node{}
	}
	if err = tree.addRouter(path, router); err != nil {
		return
	}
	m.routers[method] = tree
	return
}

func (m *goAPIMux) notFindRouters(router *appRouter) {
	m.notFind = router
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
	for _, val := range m.statics[method] {
		if strings.HasPrefix(urlPath, val.path) {
			router = val
			paths = map[string]string{}
			exists = true
			return
		}
	}
	tree := m.routers[method]
	if tree == nil {
		return
	}
	return tree.findRouter(urlPath)
}
