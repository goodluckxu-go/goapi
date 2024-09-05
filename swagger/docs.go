package swagger

import (
	"fmt"
	"net/http"
	"time"
)

type fileInfo struct {
	Path    string
	Content string
}

type Config struct {
	// label expansion mode, value in list, full, none
	DocExpansion string

	// whether to enable deep linking
	DeepLinking bool
}

type Swagger struct {
	Title                       string
	Index                       fileInfo
	CssIndex                    fileInfo
	CssSwaggerUI                fileInfo
	JsSwaggerInitializer        fileInfo
	JsSwaggerUiBundle           fileInfo
	JsSwaggerUiStandalonePreset fileInfo
	OpenAPIPath                 string
}

type Router struct {
	Path    string
	Handler func(writer http.ResponseWriter, request *http.Request)
}

func GetSwagger(path, title, favicon string, openapiJsonBody []byte, config Config) (routers []Router) {
	openapiPath := path + "/openapi.json"
	routers = append(routers, Router{
		Path: path,
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			if handleCache(writer, request) {
				return
			}
			_, _ = writer.Write([]byte(fmt.Sprintf(index, title, path, path, favicon, path, path, path)))
		},
	}, Router{
		Path: path + "/index.css",
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/css; charset=utf-8")
			if handleCache(writer, request) {
				return
			}
			_, _ = writer.Write([]byte(cssIndex))
		},
	}, Router{
		Path: path + "/swagger-ui.css",
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/css; charset=utf-8")
			if handleCache(writer, request) {
				return
			}
			_, _ = writer.Write([]byte(cssSwaggerUi))
		},
	}, Router{
		Path: path + "/swagger-initializer.js",
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			if handleCache(writer, request) {
				return
			}
			_, _ = writer.Write([]byte(fmt.Sprintf(jsSwaggerInitializer, openapiPath, config.DocExpansion, config.DeepLinking)))
		},
	}, Router{
		Path: path + "/swagger-ui-bundle.js",
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			if handleCache(writer, request) {
				return
			}
			_, _ = writer.Write([]byte(jsSwaggerUiBundle))
		},
	}, Router{
		Path: path + "/swagger-ui-standalone-preset.js",
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			if handleCache(writer, request) {
				return
			}
			_, _ = writer.Write([]byte(jsSwaggerUiStandalonePreset))
		},
	}, Router{
		Path: openapiPath,
		Handler: func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = writer.Write(openapiJsonBody)
		},
	})
	return
}

func handleCache(writer http.ResponseWriter, request *http.Request) bool {
	if request.Header.Get("If-Modified-Since") != "" {
		writer.WriteHeader(304)
		return true
	}
	writer.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	writer.Header().Set("Cache-Control", "max-age=86400")
	return false
}
