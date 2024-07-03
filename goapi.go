package goapi

import (
	"encoding/json"
	"fmt"
	"github.com/goodluckxu-go/goapi/lang"
	"github.com/goodluckxu-go/goapi/openapi"
	"github.com/goodluckxu-go/goapi/swagger"
	"net/http"
	"os"
)

// GoAPI It is a newly created API function
func GoAPI(app APP, isDocs bool, docsPath ...string) *API {
	dPath := "/docs"
	if len(docsPath) > 0 {
		dPath = docsPath[0]
	}
	return &API{
		app:    app,
		isDocs: isDocs,
		OpenAPIInfo: &openapi.Info{
			Title:   "GoAPI",
			Version: "1.0.0",
		},
		log:      &levelHandleLogger{log: &defaultLogger{}},
		docsPath: dPath,
		addr:     ":8080",
		lang:     &lang.EN{},
	}
}

type API struct {
	app                   APP
	handlers              []any
	httpExceptionResponse Response
	responseMediaTypes    []MediaType
	OpenAPIInfo           *openapi.Info
	isDocs                bool
	OpenAPIServers        []*openapi.Server
	OpenAPITags           []*openapi.Tag
	docsPath              string
	exceptFunc            func(httpCode int, detail string) Response
	lang                  Lang
	log                   Logger
	addr                  string
	routers               []appRouter
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

// Run It is an execution function
func (a *API) Run(addr ...string) error {
	if len(addr) > 0 {
		a.addr = addr[0]
	}
	a.init()
	handle := newHandler(a)
	handle.Handle()
	if a.isDocs {
		api := newHandlerOpenAPI(a, handle).Handle()
		openapiBody, _ := json.Marshal(api)
		a.app.Init()
		list := swagger.GetSwagger(a.docsPath, api.Info.Title, "", openapiBody)
		for _, v := range list {
			a.routers = append(a.routers, a.handleSwagger(v, handle.middlewares))
		}
	} else {
		a.app.Init()
	}
	newHandlerServer(a, handle).Handle()
	a.log.Info("Started server process [%v]", colorDebug(os.Getpid()))
	a.log.Info("Using the [%v] APP", colorDebug(fmt.Sprintf("%T", a.app)))
	a.log.Info("GoAPI running on http://%v (Press CTRL+C to quit)", a.addr)
	return a.app.Run(a.addr)
}

func (a *API) handleSwagger(router swagger.Router, middlewares []Middleware) appRouter {
	return appRouter{
		path:   router.Path,
		method: http.MethodGet,
		handler: func(ctx *Context) {
			ctx.middlewares = middlewares
			ctx.log = a.log
			ctx.routerFunc = func(done chan struct{}) {
				router.Handler(ctx.Writer)
				done <- struct{}{}
			}
			ctx.Next()
		},
	}
}

func (a *API) init() {
	if len(a.responseMediaTypes) == 0 {
		a.responseMediaTypes = []MediaType{JSON}
	}
	if a.exceptFunc == nil {
		a.exceptFunc = func(httpCode int, detail string) Response {
			return &HTTPResponse[any]{
				HttpCode: httpCode,
				Body:     detail,
			}
		}
		a.httpExceptionResponse = a.exceptFunc(0, "")
	}
}

type Middleware func(ctx *Context)
