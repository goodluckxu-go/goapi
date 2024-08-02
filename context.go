package goapi

import (
	"bufio"
	"net"
	"net/http"
	"sync"
	"time"
)

type Context struct {
	Request     *http.Request
	Writer      http.ResponseWriter
	Values      map[string]any
	log         Logger
	mux         sync.RWMutex
	middlewares []Middleware
	routerFunc  func(done chan struct{})
	paths       map[string]string
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

// Set It is a method for setting context values
func (c *Context) Set(key string, value any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Values == nil {
		c.Values = map[string]any{}
	}
	c.Values[key] = value
}

// Get It is a method for obtaining context values
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

// Next It is used in middleware, before Next is before interface request, and after Next is after interface request
func (c *Context) Next() {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if len(c.middlewares) == 0 {
		if c.routerFunc != nil {
			done := make(chan struct{})
			go c.routerFunc(done)
			<-done
		}
		return
	}
	middleware := c.middlewares[0]
	c.middlewares = c.middlewares[1:]
	middleware(c)
}

// Logger It is a method of obtaining logs
func (c *Context) Logger() Logger {
	return c.log
}

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

// Hijack implements the http.Hijacker interface.
func (r *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush implements the http.Flusher interface.
func (r *ResponseWriter) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

func (r *ResponseWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	return r.ResponseWriter.Write(b)
}

func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

func (r *ResponseWriter) Status() int {
	return r.status
}
