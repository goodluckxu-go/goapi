package goapi

import (
	"encoding/json"
	"github.com/goodluckxu-go/goapi/lang"
	"github.com/goodluckxu-go/goapi/openapi"
	"github.com/goodluckxu-go/goapi/swagger"
	"net/http"
)

func GoAPI(app APP, isDocs bool, docsPath ...string) *API {
	dPath := "docs"
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
		docsPath: dPath,
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
	routers               []*AppRouter
}

func (a *API) HTTPExceptionHandler(f func(httpCode int, detail string) Response) {
	a.httpExceptionResponse = f(0, "")
	a.exceptFunc = f
}

func (a *API) SetLang(lang Lang) {
	a.lang = lang
}

func (a *API) SetLogger(log Logger) {
	a.log = log
}

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

func (a *API) AddMiddleware(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		a.handlers = append(a.handlers, middleware)
	}
}

func (a *API) IncludeRouter(router any, prefix string, isDocs bool, middlewares ...Middleware) {
	a.handlers = append(a.handlers, &includeRouter{
		router:      router,
		prefix:      prefix,
		isDocs:      isDocs,
		middlewares: middlewares,
	})
}

func (a *API) Run(addr ...string) error {
	a.init()
	handle := newHandler(a)
	handle.Handle()
	if a.isDocs {
		api := newHandlerOpenAPI(a, handle.paths, handle.structs).Handle()
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

	return a.app.Run(addr...)
}

func (a *API) handleSwagger(router swagger.Router, middlewares []Middleware) *AppRouter {
	return &AppRouter{
		Path:   router.Path,
		Method: http.MethodGet,
		Handler: func(ctx *Context) {
			ctx.middlewares = middlewares
			ctx.Log = a.log
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
	if a.lang == nil {
		a.lang = &lang.EN{}
	}
	if a.log == nil {
		a.log = &defaultLogger{}
	}
}

type Middleware func(ctx *Context)
