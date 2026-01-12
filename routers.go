package goapi

import (
	"github.com/goodluckxu-go/goapi/openapi"
	"github.com/goodluckxu-go/goapi/swagger"
)

type returnObject interface {
	returnObj() (obj returnObjResult, err error)
}

type RouterInterface interface {
	RouterGroupInterface
	Child(prefix string, docsPath string) *RouterChild
}

type IRouters struct {
	RouterChild
}

// Child It is an introduction routing children
func (i *IRouters) Child(prefix string, docsPath string) *RouterChild {
	child := &RouterChild{
		RouterGroup: RouterGroup{
			prefix:      pathJoin(i.prefix, prefix),
			groupPrefix: pathJoin(i.groupPrefix, prefix),
			isDocs:      i.isDocs,
			docsPath:    pathJoin(i.docsPath, docsPath),
			middlewares: append(i.middlewares, i.getMiddlewares()...),
		},
		IsDocs: true,
		OpenAPIInfo: &openapi.Info{
			Title:   "GoAPI",
			Version: "1.0.0",
		},
		Swagger: swagger.Config{
			DocExpansion: "list",
			DeepLinking:  true,
		},
	}
	i.handlers = append(i.handlers, child)
	return child
}

func (i *IRouters) returnObj() (obj returnObjResult, err error) {
	return i.RouterChild.returnObj()
}
