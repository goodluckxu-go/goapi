package goapi

import (
	"encoding/json"
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
}

func (a *API) HTTPExceptionHandler(f func(httpCode int, detail string) Response) {
	a.httpExceptionResponse = f(0, "")
	a.exceptFunc = f
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
		api := newHandlerOpenAPI(a, handle.list, handle.structs).Handle()
		a.app.Init()
		a.swagger(a.app, api)
	} else {
		a.app.Init()
	}
	newHandlerServer(a.app, handle.list, a.exceptFunc, handle.structs).Handle()

	return a.app.Run(addr...)
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

func (a *API) swagger(app APP, api *openapi.OpenAPI) {
	jsonByte, _ := json.Marshal(api)
	swagInfo := swagger.GetSwagger(a.docsPath, api.Info.Title, "")
	app.GET(swagInfo.Index.Path, func(req *http.Request, writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(swagInfo.Index.Content))
	})
	app.GET(swagInfo.CssIndex.Path, func(req *http.Request, writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = writer.Write([]byte(swagInfo.CssIndex.Content))
	})
	app.GET(swagInfo.CssSwaggerUI.Path, func(req *http.Request, writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = writer.Write([]byte(swagInfo.CssSwaggerUI.Content))
	})
	app.GET(swagInfo.JsSwaggerInitializer.Path, func(req *http.Request, writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		_, _ = writer.Write([]byte(swagInfo.JsSwaggerInitializer.Content))
	})
	app.GET(swagInfo.JsSwaggerUiBundle.Path, func(req *http.Request, writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		_, _ = writer.Write([]byte(swagInfo.JsSwaggerUiBundle.Content))
	})
	app.GET(swagInfo.JsSwaggerUiStandalonePreset.Path, func(req *http.Request, writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		_, _ = writer.Write([]byte(swagInfo.JsSwaggerUiStandalonePreset.Content))
	})
	app.GET(swagInfo.OpenAPIPath, func(req *http.Request, writer http.ResponseWriter) {
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = writer.Write(jsonByte)
	})
}

type Middleware func(ctx *Context)
