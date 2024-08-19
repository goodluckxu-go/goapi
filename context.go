package goapi

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Context struct {
	Request     *http.Request
	Writer      http.ResponseWriter
	writermem   ResponseWriter
	Values      map[string]any
	log         Logger
	mux         sync.RWMutex
	middlewares []Middleware
	paths       map[string]string
	index       int
	fullPath    string
}

func (c *Context) reset() {
	c.Writer = &c.writermem
	c.Values = nil
	c.middlewares = c.middlewares[:]
	c.paths = nil
	c.index = -1
	c.fullPath = ""
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
	c.index++
	for ; c.index < len(c.middlewares); c.index++ {
		c.middlewares[c.index](c)
	}
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
		ip := net.ParseIP(strings.Split(xForwardedFor, ",")[0])
		if ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	if xRealIP := c.Request.Header.Get("X-Real-IP"); xRealIP != "" {
		ip := net.ParseIP(strings.Split(xRealIP, ",")[0])
		if ip != nil && ip.To4() != nil {
			return ip.String()
		}
	}
	return remoteIP.String()
}

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func (r *ResponseWriter) reset(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.status = 200
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
