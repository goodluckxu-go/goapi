package goapi

import (
	"github.com/goodluckxu-go/goapi/openapi"
)

type returnObject interface {
	returnObj() (obj returnObjResult, err error)
}

type RouterInterface interface {
	RouterGroupInterface
	Child(prefix string, isDocs bool, docsPath string) *RouterChild
}

type IRouters struct {
	RouterChild
}

// Child It is an introduction routing children
func (i *IRouters) Child(prefix string, isDocs bool, docsPath string) *RouterChild {
	child := &RouterChild{
		RouterGroup: RouterGroup{
			prefix:      pathJoin(i.prefix, prefix),
			groupPrefix: pathJoin(i.groupPrefix, prefix),
			isDocs:      i.isDocs && isDocs,
			docsPath:    pathJoin(i.docsPath, docsPath),
			middlewares: append(i.middlewares, i.getMiddlewares()...),
		},
		OpenAPIInfo: &openapi.Info{
			Title:   "GoAPI",
			Version: "1.0.0",
		},
	}
	i.handlers = append(i.handlers, child)
	return child
}

func (i *IRouters) returnObj() (obj returnObjResult, err error) {
	return i.RouterChild.returnObj()
}
