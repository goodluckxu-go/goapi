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
func (g *APIGroup) AddMiddleware(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		g.handlers = append(g.handlers, middleware)
	}
}

// IncludeRouter It is a function that introduces routing structures
func (g *APIGroup) IncludeRouter(router any, prefix string, isDocs bool, middlewares ...Middleware) {
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
