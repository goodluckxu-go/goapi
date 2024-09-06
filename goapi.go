package goapi

import (
	"fmt"
	"github.com/goodluckxu-go/goapi/lang"
	"github.com/goodluckxu-go/goapi/openapi"
	"github.com/goodluckxu-go/goapi/response"
	"github.com/goodluckxu-go/goapi/swagger"
	json "github.com/json-iterator/go"
	"net/http"
	"os"
	"strconv"
)

// GoAPI It is a newly created API function
func GoAPI(isDocs bool, docsPath ...string) *API {
	dPath := "/docs"
	if len(docsPath) > 0 {
		dPath = docsPath[0]
	}
	return &API{
		isDocs: isDocs,
		OpenAPIInfo: &openapi.Info{
			Title:   "GoAPI",
			Version: "1.0.0",
		},
		Swagger: swagger.Config{
			DocExpansion: "list",
			DeepLinking:  true,
		},
		log:      &levelHandleLogger{log: &defaultLogger{}},
		docsPath: dPath,
		addr:     ":8080",
		lang:     &lang.EnUs{},
	}
}

type API struct {
	handlers              []any
	httpExceptionResponse Response
	responseMediaTypes    []MediaType
	OpenAPIInfo           *openapi.Info
	isDocs                bool
	OpenAPIServers        []*openapi.Server
	OpenAPITags           []*openapi.Tag
	Swagger               swagger.Config
	docsPath              string
	exceptFunc            func(httpCode int, detail string) Response
	lang                  Lang
	log                   Logger
	addr                  string
	routers               []*appRouter
}

// HTTPExceptionHandler It is an exception handling registration for HTTP
func (a *API) HTTPExceptionHandler(f func(httpCode int, detail string) Response) {
	a.httpExceptionResponse = f(0, "")
	a.exceptFunc = f
}

// SetLang It is to set the validation language function
func (a *API) SetLang(lang Lang) {
	a.lang = lang
}

// SetLogger It is a function for setting custom logs
func (a *API) SetLogger(log Logger) {
	a.log = &levelHandleLogger{log: log}
}

// Logger It is a method of obtaining logs
func (a *API) Logger() Logger {
	return a.log
}

// SetResponseMediaType It is a function that sets the return value type
func (a *API) SetResponseMediaType(mediaTypes ...MediaType) {
	m := map[MediaType]struct{}{}
	for _, v := range a.responseMediaTypes {
		m[v] = struct{}{}
	}
	for _, v := range mediaTypes {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		a.responseMediaTypes = append(a.responseMediaTypes, v)
	}
}

// AddMiddleware It is a function for adding middleware
func (a *API) AddMiddleware(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		a.handlers = append(a.handlers, middleware)
	}
}

// IncludeRouter It is a function that introduces routing structures
func (a *API) IncludeRouter(router any, prefix string, isDocs bool, middlewares ...Middleware) {
	a.handlers = append(a.handlers, &includeRouter{
		router:      router,
		prefix:      prefix,
		isDocs:      isDocs,
		middlewares: middlewares,
	})
}

// IncludeGroup It is an introduction routing group
func (a *API) IncludeGroup(group *APIGroup) {
	a.handlers = append(a.handlers, group)
}

// DebugPprof Open the system's built-in pprof
func (a *API) DebugPprof() {
	a.handlers = append(a.handlers, &includeRouter{
		router: &pprofInfo{},
		prefix: "/debug",
		isDocs: true,
	})
}

// Static serves files from the given file system root.
func (a *API) Static(path, root string) {
	a.handlers = append(a.handlers, &staticInfo{
		path: path,
		root: root,
	})
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (a *API) Run(addr ...string) (err error) {
	if len(addr) > 0 {
		a.addr = addr[0]
	}
	httpHandler := a.Handler()
	a.log.Info("GoAPI running on http://%v (Press CTRL+C to quit)", a.addr)
	return http.ListenAndServe(a.addr, httpHandler)
}

// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (a *API) RunTLS(addr, certFile, keyFile string) (err error) {
	a.addr = addr
	httpHandler := a.Handler()
	a.log.Info("GoAPI running on http://%v (Press CTRL+C to quit)", a.addr)
	return http.ListenAndServeTLS(a.addr, certFile, keyFile, httpHandler)
}

// Handler Return to http.Handler interface
func (a *API) Handler() http.Handler {
	a.init()
	handle := newHandler(a)
	handle.Handle()
	if a.isDocs {
		api := newHandlerOpenAPI(a, handle).Handle()
		openapiBody, _ := json.Marshal(api)
		list := swagger.GetSwagger(a.docsPath, api.Info.Title, logo, openapiBody, a.Swagger)
		for _, v := range list {
			a.routers = append(a.routers, a.handleSwagger(v, handle.defaultMiddlewares))
		}
	}
	serverHandler := newHandlerServer(a, handle)
	serverHandler.Handle()
	pid := strconv.Itoa(os.Getpid())
	if isDefaultLogger(a.log) {
		pid = colorDebug(pid)
	}
	a.log.Info("Started server process [%v]", pid)
	a.log.Debug("All routes:")
	maxMethodLen := 0
	maxPathLen := 0
	for _, v := range a.routers {
		if maxMethodLen < len(v.method) {
			maxMethodLen = len(v.method)
		}
		if maxPathLen < len(v.path) {
			maxPathLen = len(v.path)
		}
	}
	for _, v := range a.routers {
		a.log.Debug("%v%v--> %v", spanFill(v.method, len(v.method), maxMethodLen+1), spanFill(v.path, len(v.path), maxPathLen+1), v.pos)
	}
	return serverHandler.HttpHandler()
}

func (a *API) handleSwagger(router swagger.Router, middlewares []Middleware) *appRouter {
	return &appRouter{
		path:   router.Path,
		method: http.MethodGet,
		handler: func(ctx *Context) {
			ctx.middlewares = middlewares
			ctx.log = a.log
			ctx.middlewares = append(ctx.middlewares, func(ctx *Context) {
				router.Handler(ctx.Writer, ctx.Request)
			})
			ctx.Next()
		},
		pos: fmt.Sprintf("github.com/goodluckxu-go/goapi/swagger.GetSwagger (%v Middleware)", len(middlewares)),
	}
}

func (a *API) init() {
	if len(a.responseMediaTypes) == 0 {
		a.responseMediaTypes = []MediaType{JSON}
	}
	if a.exceptFunc == nil {
		a.exceptFunc = func(httpCode int, detail string) Response {
			return &response.HTTPResponse[any]{
				HttpCode: httpCode,
				Body:     detail,
			}
		}
		a.httpExceptionResponse = a.exceptFunc(0, "")
	}
}

type Middleware func(ctx *Context)

type APIGroup struct {
	prefix   string
	isDocs   bool
	handlers []any
}

// NewGroup It is a newly created APIGroup function
func NewGroup(prefix string, isDocs bool) *APIGroup {
	return &APIGroup{
		prefix: prefix,
		isDocs: isDocs,
	}
}

// AddMiddleware It is a function for adding middleware
func (g *APIGroup) AddMiddleware(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		g.handlers = append(g.handlers, middleware)
	}
}

// IncludeRouter It is a function that introduces routing structures
func (g *APIGroup) IncludeRouter(router any, prefix string, isDocs bool, middlewares ...Middleware) {
	g.handlers = append(g.handlers, &includeRouter{
		router:      router,
		prefix:      prefix,
		isDocs:      isDocs,
		middlewares: middlewares,
	})
}

// IncludeGroup It is an introduction routing group
func (g *APIGroup) IncludeGroup(group *APIGroup) {
	g.handlers = append(g.handlers, group)
}
