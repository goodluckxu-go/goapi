package goapi

import (
	"net/http"
	"sync"
	"time"
)

type Context struct {
	Request     *http.Request
	Writer      http.ResponseWriter
	Values      map[string]any
	mux         sync.RWMutex
	middlewares []Middleware
	routerFunc  func()
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	if c.Request != nil && c.Request.Context() != nil {
		return c.Request.Context().Deadline()
	}
	return
}

func (c *Context) Done() <-chan struct{} {
	if c.Request != nil && c.Request.Context() != nil {
		return c.Request.Context().Done()
	}
	return nil
}

func (c *Context) Err() error {
	if c.Request != nil && c.Request.Context() != nil {
		return c.Request.Context().Err()
	}
	return nil
}

func (c *Context) Set(key string, value any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Values == nil {
		c.Values = map[string]any{}
	}
	c.Values[key] = value
}

func (c *Context) Get(key string) (value any, ok bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	value, ok = c.Values[key]
	return
}

func (c *Context) Value(key any) any {
	if k, ok := key.(string); ok {
		if value, keyOk := c.Get(k); keyOk {
			return value
		}
	}
	if c.Request != nil && c.Request.Context() != nil {
		return c.Request.Context().Value(key)
	}
	return nil
}

func (c *Context) Next() {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if len(c.middlewares) == 0 {
		c.routerFunc()
		return
	}
	middleware := c.middlewares[0]
	c.middlewares = c.middlewares[1:]
	middleware(c)
}
