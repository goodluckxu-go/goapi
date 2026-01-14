package goapi

import (
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

type Context struct {
	Request    *http.Request
	Writer     ResponseWriter
	writermem  responseWriter
	Values     map[string]any
	log        Logger
	mux        sync.RWMutex
	handlers   []HandleFunc
	Params     Params
	index      int
	fullPath   string
	mediaType  string
	path       *pathInfo
	queryCache url.Values
}

func (c *Context) reset() {
	c.Writer = &c.writermem
	c.Params = c.Params[:0]
	c.Values = nil
	c.handlers = c.handlers[0:0]
	c.index = -1
	c.fullPath = ""
	c.queryCache = nil
	levelLog, _ := c.log.(*levelHandleLogger)
	if _, ok := getFnByCovertInterface[LoggerRequestID](levelLog.log); ok {
		newLog := c.copyLogger(levelLog.log)
		if fn, fnOk := newLog.(LoggerRequestID); fnOk {
			pk, _ := uuid.NewV4()
			fn.SetRequestID(pk.String())
		}
		c.log = &levelHandleLogger{log: newLog}
	}
}

func (c *Context) copyLogger(log Logger) Logger {
	val := reflect.ValueOf(log)
	var newVal reflect.Value
	if val.Kind() == reflect.Ptr {
		newVal = reflect.New(val.Type().Elem())
	} else {
		newVal = reflect.New(val.Type()).Elem()
	}
	c.copyStruct(newVal, val)
	return newVal.Interface().(Logger)
}

func (c *Context) copyStruct(dst, src reflect.Value) {
	if dst.Type() != src.Type() || src.IsZero() {
		return
	}
	switch src.Kind() {
	case reflect.Ptr:
		c.copyStruct(dst.Elem(), src.Elem())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		dst.Set(src)
	case reflect.Slice, reflect.Array:
		for i := 0; i < src.Len(); i++ {
			c.copyStruct(dst.Index(i), src.Index(i))
		}
	case reflect.Map:
		keys := src.MapKeys()
		for _, key := range keys {
			dst.SetMapIndex(key, src.MapIndex(key))
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			c.copyStruct(dst.Field(i), src.Field(i))
		}
	default:
	}
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
