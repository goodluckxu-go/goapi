package goapi

import (
	"net/http"

	"github.com/goodluckxu-go/goapi/v2/openapi"
	"github.com/goodluckxu-go/goapi/v2/swagger"
)

type RouterChildInterface interface {
	RouterGroupInterface
	HTTPException(handler func(httpCode int, detail string) any)
	NoRoute(handler func(ctx *Context))
	NoMethod(handler func(ctx *Context))
}

type RouterGroupInterface interface {
	AddMiddleware(middlewares ...HandleFunc)
	IncludeRouter(router any, prefix string, isDocs bool, middlewares ...HandleFunc)
	Group(prefix string, isDocs bool) *RouterGroup
	StaticFile(path, root string)
	Static(path, root string)
	StaticFS(path string, fs http.FileSystem)
}

type RouterChild struct {
	RouterGroup
	IsDocs                 bool // default true
	OpenAPIInfo            *openapi.Info
	OpenAPIServers         []*openapi.Server
	OpenAPITags            []*openapi.Tag
	Swagger                swagger.Config
	RedirectTrailingSlash  bool
	HandleMethodNotAllowed bool // support http.StatusMethodNotAllowed
	// func set
	noRoute    func(ctx *Context)
	noMethod   func(ctx *Context)
	exceptFunc func(httpCode int, detail string) any
}

// HTTPException adds handlers for http exception
func (r *RouterChild) HTTPException(handler func(httpCode int, detail string) any) {
	r.exceptFunc = handler
}

// NoRoute adds handlers for NoRoute. It returns a 404 code by default.
func (r *RouterChild) NoRoute(handler func(ctx *Context)) {
	r.noRoute = handler
}

// NoMethod sets the handlers called when HandleMethodNotAllowed = true.
func (r *RouterChild) NoMethod(handler func(ctx *Context)) {
	r.noMethod = handler
}

func (r *RouterChild) init() *RouterChild {
	r.IsDocs = true
	r.OpenAPIInfo = &openapi.Info{
		Title:   "GoAPI",
		Version: "1.0.0",
	}
	r.Swagger = swagger.Config{
		DocExpansion: "list",
		DeepLinking:  true,
	}
	r.RedirectTrailingSlash = true
	r.noRoute = defaultNoRoute
	r.noMethod = defaultNoMethod
	r.exceptFunc = defaultExceptFunc
	return r
}

func (r *RouterChild) returnObj() (obj returnObjResult, err error) {
	obj, err = r.RouterGroup.returnObj()
	if err != nil {
		return
	}
	for _, path := range obj.paths {
		if path.docsPath == r.docsPath {
			path.isDocs = path.isDocs && r.IsDocs
		}
	}
	docs := obj.docsMap[r.docsPath]
	docs.isDocs = r.IsDocs
	docs.info = r.OpenAPIInfo
	docs.servers = r.OpenAPIServers
	docs.tags = mergeOpenAPITags(docs.tags, r.OpenAPITags)
	docs.swagger = r.Swagger
	obj.docsMap[r.docsPath] = docs
	child := obj.childMap[r.childPath]
	child.redirectTrailingSlash = r.RedirectTrailingSlash
	child.handleMethodNotAllowed = r.HandleMethodNotAllowed
	child.noRoute = r.noRoute
	child.noMethod = r.noMethod
	child.exceptFunc = r.exceptFunc
	obj.childMap[r.childPath] = child
	return
}

type RouterGroup struct {
	prefix      string
	groupPrefix string
	isDocs      bool
	docsPath    string
	childPath   string
	middlewares []HandleFunc
	handlers    []any
}

// AddMiddleware It is a function for adding middleware
func (r *RouterGroup) AddMiddleware(middlewares ...HandleFunc) {
	for _, middleware := range middlewares {
		r.handlers = append(r.handlers, middleware)
	}
}

// IncludeRouter It is a function that introduces routing structures
func (r *RouterGroup) IncludeRouter(router any, prefix string, isDocs bool, middlewares ...HandleFunc) {
	r.handlers = append(r.handlers, &includeRouter{
		router:      router,
		prefix:      pathJoin(r.prefix, prefix),
		groupPrefix: r.groupPrefix,
		isDocs:      r.isDocs && isDocs,
		docsPath:    r.docsPath,
		childPath:   r.childPath,
		middlewares: append(r.middlewares, append(r.getMiddlewares(), middlewares...)...),
	})
}

// StaticFile registers a single route in order to serve a single file of the local filesystem.
// router.StaticFile("favicon.ico", "./resources/favicon.ico")
func (r *RouterGroup) StaticFile(path, root string) {
	r.handlers = append(r.handlers, &staticInfo{
		path:        path,
		fs:          http.Dir(root),
		isFile:      true,
		groupPrefix: r.groupPrefix,
		middlewares: append(r.middlewares, r.getMiddlewares()...),
	})
}

// Static serves files from the given file system root.
func (r *RouterGroup) Static(path, root string) {
	r.StaticFS(path, Dir(root, false))
}

// StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
// goapi by default uses: goapi.Dir()
func (r *RouterGroup) StaticFS(path string, fs http.FileSystem) {
	r.handlers = append(r.handlers, &staticInfo{
		path:        path,
		fs:          fs,
		groupPrefix: r.groupPrefix,
		middlewares: append(r.middlewares, r.getMiddlewares()...),
	})
}

// Group It is an introduction routing group
func (r *RouterGroup) Group(prefix string, isDocs bool) *RouterGroup {
	group := &RouterGroup{
		prefix:      pathJoin(r.prefix, prefix),
		groupPrefix: pathJoin(r.groupPrefix, prefix),
		isDocs:      r.isDocs && isDocs,
		docsPath:    r.docsPath,
		childPath:   r.childPath,
		middlewares: append(r.middlewares, r.getMiddlewares()...),
	}
	r.handlers = append(r.handlers, group)
	return group
}

func (r *RouterGroup) getMiddlewares() (middlewares []HandleFunc) {
	for _, hd := range r.handlers {
		if middleware, ok := hd.(HandleFunc); ok {
			middlewares = append(middlewares, middleware)
		}
	}
	return middlewares
}

func (r *RouterGroup) returnObj() (obj returnObjResult, err error) {
	obj.groupMap = map[string]returnObjGroup{
		r.prefix: {
			middlewares: r.getMiddlewares(),
		},
	}
	obj.docsMap = map[string]returnObjDocs{}
	obj.childMap = map[string]returnObjChild{}
	obj.mediaTypes = map[MediaType]struct{}{}
	var childObj returnObjResult
	for _, hd := range r.handlers {
		if fn, ok := hd.(returnObject); ok {
			childObj, err = fn.returnObj()
			if err != nil {
				return
			}
			for k, v := range childObj.groupMap {
				if k == r.groupPrefix {
					v.middlewares = append(obj.groupMap[k].middlewares, v.middlewares...)
				}
				obj.groupMap[k] = v
			}
			for k, v := range childObj.docsMap {
				v.tags = mergeOpenAPITags(obj.docsMap[k].tags, v.tags)
				obj.docsMap[k] = v
			}
			for k, v := range childObj.childMap {
				obj.childMap[k] = v
			}
			for k, v := range childObj.mediaTypes {
				obj.mediaTypes[k] = v
			}
			obj.paths = append(obj.paths, childObj.paths...)
		}
	}
	return
}
