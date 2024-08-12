package goapi

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func newGoAPIMux(log Logger) *goAPIMux {
	return &goAPIMux{
		routers: map[string]*routerPath{},
		log:     log,
	}
}

type goAPIMux struct {
	routers map[string]*routerPath
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
	rPath := m.routers[method]
	if rPath == nil {
		rPath = &routerPath{
			statics: map[string]*appRouter{},
		}
	}
	if err = rPath.addPaths(path, router); err != nil {
		return
	}
	m.routers[method] = rPath
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
	router.handler(ctx)
}

func (m *goAPIMux) searchRouters(urlPath, method string) (router *appRouter, paths map[string]string, exists bool) {
	rPath := m.routers[method]
	if rPath == nil {
		return
	}
	return rPath.search(urlPath)
}

type routerPath struct {
	statics map[string]*appRouter
	matches []*routerMatch
}

func (r *routerPath) addPaths(path string, router *appRouter) (err error) {
	left := strings.Index(path, "{")
	right := strings.Index(path, "}")
	if left == -1 && right == -1 && !router.isPrefix {
		r.statics[path] = router
		return
	}
	match := &routerMatch{
		router: router,
	}
	if router.isPrefix {
		match.prefix = path
		r.matches = append(r.matches, match)
		return
	}
	for {
		if left == -1 && right == -1 {
			break
		}
		if (left == -1 && right != -1) || (left != -1 && right == -1) || left > right {
			err = fmt.Errorf("path format error")
			return
		}
		fixed := path[:left]
		param := path[left+1 : right]
		if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(param) {
			err = fmt.Errorf("path format error")
			return
		}
		path = path[right+1:]
		match.fixedSeg = append(match.fixedSeg, fixed)
		match.params = append(match.params, param)
		left = strings.Index(path, "{")
		right = strings.Index(path, "}")
	}
	if path != "" {
		match.fixedSeg = append(match.fixedSeg, path)
	}
	r.matches = append(r.matches, match)
	return
}

func (r *routerPath) search(urlPath string) (router *appRouter, paths map[string]string, exists bool) {
	if router = r.statics[urlPath]; router != nil {
		exists = true
		return
	}
	for _, match := range r.matches {
		if paths, exists = match.matchValue(urlPath); exists {
			router = match.router
			return
		}
	}
	return
}

type routerMatch struct {
	fixedSeg []string
	params   []string
	prefix   string
	router   *appRouter
}

func (r *routerMatch) matchValue(urlPath string) (values map[string]string, exists bool) {
	if r.prefix != "" && strings.HasPrefix(urlPath, r.prefix) {
		exists = true
		return
	}
	fixLeft := 0
	paramLeft := -1
	values = map[string]string{}
	for fixLeft < len(r.fixedSeg) && paramLeft < len(r.params) {
		idx := strings.Index(urlPath, r.fixedSeg[fixLeft])
		if idx == -1 {
			values = nil
			return
		}
		if paramLeft != -1 {
			values[r.params[paramLeft]] = urlPath[:idx]
		}
		urlPath = urlPath[idx+len(r.fixedSeg[fixLeft]):]
		fixLeft++
		paramLeft++
	}
	if fixLeft < len(r.fixedSeg) {
		idx := strings.Index(urlPath, r.fixedSeg[fixLeft])
		if idx == -1 {
			values = nil
			return
		}
		values[r.params[paramLeft]] = urlPath[:idx]
		fixLeft++
		paramLeft++
	} else if paramLeft < len(r.params) && paramLeft != -1 {
		values[r.params[paramLeft]] = urlPath
		paramLeft++
	}
	if fixLeft == len(r.fixedSeg) && paramLeft == len(r.params) {
		exists = true
		return
	}
	values = nil
	return
}
