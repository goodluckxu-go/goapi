package swagger

import (
	"fmt"
	"net/http"
)

type fileInfo struct {
	Path    string
	Content string
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
	Handler func(writer http.ResponseWriter)
}

func GetSwagger(path, title, favicon string, openapiJsonBody []byte) (routers []Router) {
	openapiPath := path + "/openapi.json"
	routers = append(routers, Router{
		Path: path,
		Handler: func(writer http.ResponseWriter) {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte(fmt.Sprintf(index, title, path, path, favicon, path, path, path)))
		},
	}, Router{
		Path: path + "/index.css",
		Handler: func(writer http.ResponseWriter) {
			writer.Header().Set("Content-Type", "text/css; charset=utf-8")
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte(cssIndex))
		},
	}, Router{
		Path: path + "/swagger-ui.css",
		Handler: func(writer http.ResponseWriter) {
			writer.Header().Set("Content-Type", "text/css; charset=utf-8")
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte(cssSwaggerUi))
		},
	}, Router{
		Path: path + "/swagger-initializer.js",
		Handler: func(writer http.ResponseWriter) {
			writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte(fmt.Sprintf(jsSwaggerInitializer, openapiPath)))
		},
	}, Router{
		Path: path + "/swagger-ui-bundle.js",
		Handler: func(writer http.ResponseWriter) {
			writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte(jsSwaggerUiBundle))
		},
	}, Router{
		Path: path + "/swagger-ui-standalone-preset.js",
		Handler: func(writer http.ResponseWriter) {
			writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte(jsSwaggerUiStandalonePreset))
		},
	}, Router{
		Path: openapiPath,
		Handler: func(writer http.ResponseWriter) {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			writer.WriteHeader(200)
			_, _ = writer.Write(openapiJsonBody)
		},
	})
	return
}
