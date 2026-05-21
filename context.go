package goapi

import (
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Context struct {
	Request      *http.Request
	Writer       ResponseWriter
	writermem    responseWriter
	Values       map[string]any
	log          Logger
	mux          sync.RWMutex
	handlers     []HandleFunc
	Params       *Params
	skippedNodes *[]skippedNode
	index        int
	fullPath     string
	path         *pathInfo
	queryCache   url.Values
	ChildPath    string
	RequestID    string
	handleError  func(ctx *Context, err error)
	isRedirect   bool
	langInfo     Lang
	// prefix has 'x-'
	Extensions *Extensions
}

func (c *Context) reset() {
	c.Writer = &c.writermem
	*c.Params = (*c.Params)[:0]
	*c.skippedNodes = (*c.skippedNodes)[:0]
	// Reusing the map reduces the GC pressure and only clears without setting nil
	if c.Values != nil {
		for k := range c.Values {
			delete(c.Values, k)
		}
	}
	c.handlers = c.handlers[0:0]
	c.index = -1
	c.fullPath = ""
	c.queryCache = nil
	c.ChildPath = ""
	c.RequestID = ""
	c.isRedirect = false
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

// FullPath returns a matched route full path. For not found routes
// returns an empty string.
func (c *Context) FullPath() string {
	return c.fullPath
}

// Next It is used in middleware, before Next is before interface request, and after Next is after interface request
func (c *Context) Next() {
	defer func() {
		if err := recover(); err != nil {
			c.Logger().Fatal("panic: %v [recovered]\n%v", err, string(debug.Stack()))
			c.handleError(c, NewHTTPError(http.StatusInternalServerError, toString(err)))
		}
	}()
	c.index++
	if len(c.handlers) <= c.index {
		return
	}
	handle := c.handlers[c.index]
	if handle == nil {
		c.Next()
		return
	}
	handle(c)
}

// Copy returns a copy of the current context that can be safely used outside the request's scope.
// This has to be used when the context has to be passed to a goroutine.
func (c *Context) Copy() *Context {
	cp := Context{
		Request:     c.Request,
		log:         c.log,
		fullPath:    c.fullPath,
		queryCache:  c.queryCache,
		ChildPath:   c.ChildPath,
		RequestID:   c.RequestID,
		handleError: c.handleError,
		langInfo:    c.langInfo,
		Extensions:  c.Extensions.Root(),
	}

	cp.writermem.ResponseWriter = nil
	cp.Writer = &cp.writermem

	cp.Values = make(map[string]any, len(c.Values))
	c.mux.RLock()
	for k, v := range c.Values {
		cp.Values[k] = v
	}
	c.mux.RUnlock()

	if c.Params != nil && len(*c.Params) > 0 {
		cParams := make([]Param, len(*c.Params))
		copy(cParams, *c.Params)
		params := Params(cParams)
		cp.Params = &params
	}

	return &cp
}

// Logger It is a method of obtaining logs
func (c *Context) Logger() Logger {
	return c.log
}

// RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the IP (without the port).
func (c *Context) RemoteIP() string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

// ClientIP implements one best effort algorithm to return the real client IP.
// It is it will then try to parse the headers defined in http.Header (defaulting to [X-Forwarded-For, X-Real-Ip]).
// else the remote IP (coming from Request.RemoteAddr) is returned.
func (c *Context) ClientIP() string {
	remoteIP := net.ParseIP(c.RemoteIP())
	if remoteIP == nil {
		return ""
	}
	if xForwardedFor := c.Request.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		xForwardedFor, _, _ = strings.Cut(xForwardedFor, ",")
		ip := net.ParseIP(xForwardedFor)
		if ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	if xRealIP := c.Request.Header.Get("X-Real-IP"); xRealIP != "" {
		xRealIP, _, _ = strings.Cut(xRealIP, ",")
		ip := net.ParseIP(xRealIP)
		if ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	return remoteIP.String()
}

func (c *Context) initQueryCache() {
	if c.queryCache == nil {
		if c.Request != nil {
			c.queryCache = c.Request.URL.Query()
		} else {
			c.queryCache = url.Values{}
		}
	}
}

// Query get all Query, values can be cached
func (c *Context) Query() url.Values {
	c.initQueryCache()
	return c.queryCache
}

// Redirect returns an HTTP redirect to the specific location.
func (c *Context) Redirect(status int, location string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if !c.isRedirect {
		http.Redirect(c.Writer, c.Request, location, status)
		c.isRedirect = true
	}
}

func (c *Context) lang() Lang {
	return c.langInfo
}
