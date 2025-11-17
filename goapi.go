package goapi

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/goodluckxu-go/goapi/lang"
	"github.com/goodluckxu-go/goapi/openapi"
	"github.com/goodluckxu-go/goapi/response"
	"github.com/goodluckxu-go/goapi/swagger"
)

// GoAPI It is a newly created API function
func GoAPI(isDocs bool, docsPath ...string) *API {
	dPath := "/docs"
	if len(docsPath) > 0 {
		dPath = docsPath[0]
	}
	return &API{
		responseMediaTypes: []MediaType{JSON},
		isDocs:             isDocs,
		OpenAPIInfo: &openapi.Info{
			Title:   "GoAPI",
			Version: "1.0.0",
		},
		Swagger: swagger.Config{
			DocExpansion: "list",
			DeepLinking:  true,
		},
		exceptFunc: func(httpCode int, detail string) Response {
			return &response.HTTPResponse[string]{
				HttpCode: httpCode,
				Body:     detail,
			}
		},
		log:                  &levelHandleLogger{log: &defaultLogger{}},
		docsPath:             dPath,
		addr:                 ":8080",
		lang:                 &lang.EnUs{},
		structTagVariableMap: map[string]any{},
	}
}

type API struct {
	handlers             []any
	responseMediaTypes   []MediaType
	OpenAPIInfo          *openapi.Info
	isDocs               bool
	OpenAPIServers       []*openapi.Server
	OpenAPITags          []*openapi.Tag
	Swagger              swagger.Config
	docsPath             string
	exceptFunc           func(httpCode int, detail string) Response
	lang                 Lang
	log                  Logger
	addr                 string
	structTagVariableMap map[string]any
	autoTagsIndex        *int
}

// HTTPExceptionHandler It is an exception handling registration for HTTP
func (a *API) HTTPExceptionHandler(f func(httpCode int, detail string) Response) {
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

// SetStructTagVariableMapping It is set struct tag variable mapping
// Only supports the replacement of tags 'summary' and 'desc'
// example:
//
//	mapping:
//		sign: test
//		username: {{sign}}
//		password: {{sign}}123456
//	tag value:
//		username is {{username}} and password is {{password}}
//	result value:
//		username is test and password is test123456
func (a *API) SetStructTagVariableMapping(m map[string]string) {
	for k, v := range m {
		n := len(k)
		for i := 0; i < n; i++ {
			if k[i] == '{' || k[i] == '}' {
				log.Fatal("the struct tag variable mapping key cannot be within '{', '}'")
			}
		}
		a.structTagVariableMap[k] = v
	}
}

// SetAutoTags This is the method of automatically setting tags
// 'index' Is the index of an array that divides routing by ‘/’
// If tags are set in the route, it becomes invalid
func (a *API) SetAutoTags(index uint) {
	a.autoTagsIndex = toPtr(int(index))
}

// AddMiddleware It is a function for adding middleware
func (a *API) AddMiddleware(middlewares ...HandleFunc) {
	for _, middleware := range middlewares {
		a.handlers = append(a.handlers, middleware)
	}
}

// IncludeRouter It is a function that introduces routing structures
func (a *API) IncludeRouter(router any, prefix string, isDocs bool, middlewares ...HandleFunc) {
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

// IncludeChildAPI It is an introduction routing children
func (a *API) IncludeChildAPI(child *ChildAPI) {
	a.handlers = append(a.handlers, child)
}

// DebugPprof Open the system's built-in pprof
func (a *API) DebugPprof() {
	a.handlers = append(a.handlers, &includeRouter{
		router: &pprofInfo{},
		prefix: "/debug",
		isDocs: false,
	})
}

// StaticFile registers a single route in order to serve a single file of the local filesystem.
// router.StaticFile("favicon.ico", "./resources/favicon.ico")
func (a *API) StaticFile(path, root string) {
	a.handlers = append(a.handlers, &staticInfo{
		path:   path,
		fs:     http.Dir(root),
		isFile: true,
	})
}

// Static serves files from the given file system root.
func (a *API) Static(path, root string) {
	a.StaticFS(path, Dir(root, false))
}

// StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
// goapi by default uses: goapi.Dir()
func (a *API) StaticFS(path string, fs http.FileSystem) {
	a.handlers = append(a.handlers, &staticInfo{
		path: path,
		fs:   fs,
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
	a.log.Info("GoAPI running on http://%v (Press CTRL+C to quit)", a.printAddr(a.addr))
	return http.ListenAndServe(a.addr, httpHandler)
}

// RunTLS attaches the router to a http.Server and starts listening and serving HTTPS (secure) requests.
// It is a shortcut for http.ListenAndServeTLS(addr, certFile, keyFile, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (a *API) RunTLS(addr, certFile, keyFile string) (err error) {
	a.addr = addr
	httpHandler := a.Handler()
	a.log.Info("GoAPI running on https://%v (Press CTRL+C to quit)", a.printAddr(a.addr))
	return http.ListenAndServeTLS(a.addr, certFile, keyFile, httpHandler)
}

// Handler Return to http.Handler interface
func (a *API) Handler() http.Handler {
	pid := strconv.Itoa(os.Getpid())
	if isDefaultLogger(a.log) {
		pid = colorDebug(pid)
	}
	a.log.Info("Started server process [%v]", pid)
	handle := newHandler(a)
	handle.Handle()
	openapiHandle := newHandlerOpenAPI(a, handle)
	openapiMap := openapiHandle.Handle()
	serverHandle := newHandlerServer(handle, a.log)
	serverHandle.HandleSwagger(swagger.GetSwagger, a, openapiMap)
	serverHandle.Handle()
	return serverHandle
}

func (a *API) printAddr(addr string) string {
	addrList := strings.Split(addr, ":")
	if len(addrList) != 2 || (addrList[0] != "" && addrList[0] != "0.0.0.0") {
		return addr
	}
	return GetLocalIP() + ":" + addrList[1]
}
