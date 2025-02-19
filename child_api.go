package goapi

import "github.com/goodluckxu-go/goapi/openapi"

type ChildAPI struct {
	prefix         string
	isDocs         bool
	docsPath       string
	OpenAPIInfo    *openapi.Info
	OpenAPIServers []*openapi.Server
	OpenAPITags    []*openapi.Tag
	handlers       []any
}

// NewChildAPI It is a newly created ChildAPI function
func NewChildAPI(prefix string, isDocs bool, docsPath string) *ChildAPI {
	return &ChildAPI{
		prefix:   prefix,
		isDocs:   isDocs,
		docsPath: docsPath,
		OpenAPIInfo: &openapi.Info{
			Title:   "GoAPI",
			Version: "1.0.0",
		},
	}
}

// AddMiddleware It is a function for adding middleware
func (c *ChildAPI) AddMiddleware(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		c.handlers = append(c.handlers, middleware)
	}
}

// IncludeRouter It is a function that introduces routing structures
func (c *ChildAPI) IncludeRouter(router any, prefix string, isDocs bool, middlewares ...Middleware) {
	c.handlers = append(c.handlers, &includeRouter{
		router:      router,
		prefix:      prefix,
		isDocs:      isDocs,
		middlewares: middlewares,
	})
}

// IncludeGroup It is an introduction routing group
func (c *ChildAPI) IncludeGroup(group *APIGroup) {
	c.handlers = append(c.handlers, group)
}
