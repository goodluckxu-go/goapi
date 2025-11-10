package goapi

import (
	"github.com/goodluckxu-go/goapi/openapi"
)

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

func (c *ChildAPI) returnObj(prefix, docsPath, groupPrefix string, middlewares []Middleware, isDocs bool) (obj pathInterfaceResult, err error) {
	obj.publicMiddlewares = map[string][]Middleware{}
	obj.mediaTypes = map[MediaType]struct{}{}
	groupPrefix = pathJoin(groupPrefix, c.prefix)
	c.docsPath = pathJoin(docsPath, c.docsPath)
	c.prefix = pathJoin(prefix, c.prefix)
	c.isDocs = isDocs && c.isDocs
	obj.openapiMap = map[string]*openapi.OpenAPI{
		c.docsPath: {
			Info:    c.OpenAPIInfo,
			Servers: c.OpenAPIServers,
			Tags:    c.OpenAPITags,
		},
	}
	var childObj pathInterfaceResult
	var publicMiddlewares []Middleware
	for _, hd := range c.handlers {
		if handle, ok := hd.(pathInterface); ok {
			childObj, err = handle.returnObj(c.prefix, c.docsPath, groupPrefix, append(middlewares, publicMiddlewares...), c.isDocs)
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
		if publicMiddleware, ok := hd.(Middleware); ok {
			publicMiddlewares = append(publicMiddlewares, publicMiddleware)
		}
		obj.paths = append(obj.paths, childObj.paths...)
	}
	obj.publicMiddlewares[c.prefix] = publicMiddlewares
	obj.openapiMap[c.docsPath].Tags = mergeOpenAPITags(obj.tags, obj.openapiMap[c.docsPath].Tags)
	return
}
