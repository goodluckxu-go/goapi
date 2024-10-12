package swagger

import (
	"fmt"
	"net/http"
	"strings"
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
	routers = []Router{
		{
			Path: path,
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "text/html; charset=utf-8")
				if handleCache(writer, request) {
					return
				}
				_, _ = writer.Write([]byte(fmt.Sprintf(index, title, path, path, favicon, path, path, path)))
			},
		},
		{
			Path: path + "/{path}",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				switch strings.TrimPrefix(request.URL.Path, path) {
				case "/index.css":
					writer.Header().Set("Content-Type", "text/css; charset=utf-8")
					if handleCache(writer, request) {
						return
					}
					_, _ = writer.Write([]byte(cssIndex))
				case "/swagger-ui.css":
					writer.Header().Set("Content-Type", "text/css; charset=utf-8")
					if handleCache(writer, request) {
						return
					}
					_, _ = writer.Write([]byte(cssSwaggerUi))
				case "/swagger-initializer.js":
					writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
					if handleCache(writer, request) {
						return
					}
					_, _ = writer.Write([]byte(fmt.Sprintf(jsSwaggerInitializer, openapiPath, config.DocExpansion, config.DeepLinking)))
				case "/swagger-ui-bundle.js":
					writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
					if handleCache(writer, request) {
						return
					}
					_, _ = writer.Write([]byte(jsSwaggerUiBundle))
				case "/swagger-ui-standalone-preset.js":
					writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
					if handleCache(writer, request) {
						return
					}
					_, _ = writer.Write([]byte(jsSwaggerUiStandalonePreset))
				case "/openapi.json":
					writer.Header().Set("Content-Type", "application/json; charset=utf-8")
					_, _ = writer.Write(openapiJsonBody)
				default:
					http.NotFound(writer, request)
				}
			},
		},
	}
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
