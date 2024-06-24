package swagger

import "fmt"

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

func GetSwagger(path, title, favicon string) Swagger {
	openapiPath := path + "/openapi.json"
	return Swagger{
		Title: title,
		Index: fileInfo{
			Path:    path,
			Content: fmt.Sprintf(index, title, path, path, favicon, path, path, path),
		},
		CssIndex: fileInfo{
			Path:    path + "/index.css",
			Content: cssIndex,
		},
		CssSwaggerUI: fileInfo{
			Path:    path + "/swagger-ui.css",
			Content: cssSwaggerUi,
		},
		JsSwaggerInitializer: fileInfo{
			Path:    path + "/swagger-initializer.js",
			Content: fmt.Sprintf(jsSwaggerInitializer, openapiPath),
		},
		JsSwaggerUiBundle: fileInfo{
			Path:    path + "/swagger-ui-bundle.js",
			Content: jsSwaggerUiBundle,
		},
		JsSwaggerUiStandalonePreset: fileInfo{
			Path:    path + "/swagger-ui-standalone-preset.js",
			Content: jsSwaggerUiStandalonePreset,
		},
		OpenAPIPath: openapiPath,
	}
}
