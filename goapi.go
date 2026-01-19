package goapi

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/goodluckxu-go/goapi/v2/lang"
	"github.com/goodluckxu-go/goapi/v2/openapi"
	"github.com/goodluckxu-go/goapi/v2/swagger"
)

// GoAPI It is a newly created API function
func GoAPI(isDocs bool, docsPath ...string) *API {
	dPath := "/docs"
	if len(docsPath) > 0 {
		dPath = docsPath[0]
	}
	api := &API{
		log:                  &levelHandleLogger{log: &defaultLogger{}},
		addr:                 ":8080",
		lang:                 &lang.EnUs{},
		structTagVariableMap: map[string]any{},
		defaultMiddlewares:   []HandleFunc{setLogger()},
	}
	api.IsDocs = true
	api.OpenAPIInfo = &openapi.Info{
		Title:   "GoAPI",
		Version: "1.0.0",
	}
	api.Swagger = swagger.Config{
		DocExpansion: "list",
		DeepLinking:  true,
	}
	api.RedirectTrailingSlash = true
	api.NoRoute = defaultNoRoute
	api.NoMethod = defaultNoMethod
	api.exceptFunc = defaultExceptFunc
	api.isDocs = isDocs
	api.docsPath = dPath
	api.AddMiddleware(api.defaultMiddlewares...)
	return api
}

type API struct {
	IRouters
	defaultMiddlewares   []HandleFunc
	responseMediaTypes   []MediaType
	lang                 Lang
	log                  Logger
	addr                 string
	structTagVariableMap map[string]any
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

// DebugPprof Open the system's built-in pprof
func (a *API) DebugPprof() {
	a.IncludeRouter(&pprofInfo{}, "/debug", false)
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
	serverHandle := newHandlerServer(handle, a.log)
	if a.isDocs {
		openapiHandle := newHandlerOpenAPI(a, handle)
		openapiMap := openapiHandle.Handle()
		serverHandle.HandleSwagger(swagger.GetSwagger, openapiMap)
	}
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
