package goapi

type APIGroup struct {
	prefix   string
	isDocs   bool
	handlers []any
}

// NewGroup It is a newly created APIGroup function
func NewGroup(prefix string, isDocs bool) *APIGroup {
	return &APIGroup{
		prefix: prefix,
		isDocs: isDocs,
	}
}

// AddMiddleware It is a function for adding middleware
func (g *APIGroup) AddMiddleware(middlewares ...HandleFunc) {
	for _, middleware := range middlewares {
		g.handlers = append(g.handlers, middleware)
	}
}

// IncludeRouter It is a function that introduces routing structures
func (g *APIGroup) IncludeRouter(router any, prefix string, isDocs bool, middlewares ...HandleFunc) {
	g.handlers = append(g.handlers, &includeRouter{
		router:      router,
		prefix:      prefix,
		isDocs:      isDocs,
		middlewares: middlewares,
	})
}

// IncludeGroup It is an introduction routing group
func (g *APIGroup) IncludeGroup(group *APIGroup) {
	g.handlers = append(g.handlers, group)
}

func (g *APIGroup) returnObj(prefix, docsPath, groupPrefix string, middlewares []HandleFunc, isDocs bool) (obj pathInterfaceResult, err error) {
	obj.publicMiddlewares = map[string][]HandleFunc{}
	obj.mediaTypes = map[MediaType]struct{}{}
	groupPrefix = pathJoin(groupPrefix, g.prefix)
	g.prefix = pathJoin(prefix, g.prefix)
	g.isDocs = isDocs && g.isDocs
	var childObj pathInterfaceResult
	var publicMiddlewares []HandleFunc
	for _, hd := range g.handlers {
		if handle, ok := hd.(pathInterface); ok {
			childObj, err = handle.returnObj(g.prefix, docsPath, groupPrefix, append(middlewares, publicMiddlewares...), g.isDocs)
			if err != nil {
				return
			}
			for k, v := range childObj.publicMiddlewares {
				obj.publicMiddlewares[k] = append(obj.publicMiddlewares[k], v...)
			}
			for mediaTypes := range childObj.mediaTypes {
				obj.mediaTypes[mediaTypes] = struct{}{}
			}
			obj.tags = mergeOpenAPITags(obj.tags, childObj.tags)
		}
		if publicMiddleware, ok := hd.(HandleFunc); ok {
			publicMiddlewares = append(publicMiddlewares, publicMiddleware)
		}
		obj.paths = append(obj.paths, childObj.paths...)
	}
	obj.publicMiddlewares[g.prefix] = publicMiddlewares
	return
}
