package goapi

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
			childPath:   pathJoin(i.childPath, prefix),
			middlewares: append(i.middlewares, i.getMiddlewares()...),
		},
	}
	child.init()
	i.handlers = append(i.handlers, child)
	return child
}

func (i *IRouters) returnObj() (obj returnObjResult, err error) {
	return i.RouterChild.returnObj()
}
