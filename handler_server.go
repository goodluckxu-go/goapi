package goapi

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/goodluckxu-go/goapi/v2/openapi"
	"github.com/goodluckxu-go/goapi/v2/swagger"
	"github.com/shopspring/decimal"
)

func newHandlerServer(handle *handler, log Logger) *handlerServer {
	hs := &handlerServer{
		log:    log,
		trees:  make(methodTrees, 0, 8),
		handle: handle,
	}
	hs.pool.New = func() any {
		params := make(Params, 0, hs.maxParams)
		skippedNodes := make([]skippedNode, 0, hs.maxSkippedNodes)
		return &Context{
			Params:       &params,
			skippedNodes: &skippedNodes,
		}
	}
	return hs
}

type handlerServer struct {
	log             Logger
	handle          *handler
	trees           methodTrees
	pool            sync.Pool
	regexpCache     sync.Map // map[string]*regexp.Regexp, Cache compiled regular expressions to avoid repeated compilation
	maxParams       uint16
	maxSkippedNodes uint16
}

func (h *handlerServer) Handle() {
	for _, path := range h.handle.paths {
		h.handlePath(path)
	}
	debugPrintRouter(h.log, h.handle.paths)
}

func (h *handlerServer) HandleSwagger(openapiMap map[string]*openapi.OpenAPI) {
	pos := runtime.FuncForPC(reflect.ValueOf(swagger.GetSwagger).Pointer()).Name()
	for docsPath, openAPI := range openapiMap {
		if err := openAPI.Validate(); err != nil {
			log.Fatal(err)
		}
		openapiBody, _ := json.Marshal(openAPI)
		routers := swagger.GetSwagger(docsPath, openAPI.Info.Title, openapiBody, h.handle.swaggerMap[docsPath])
		for _, router := range routers {
			h.handleSwagger(router, pos)
		}
	}
}

func (h *handlerServer) handleSwagger(router swagger.Router, pos string) {
	middlewares := h.getMiddlewares(router.Paths[0])
	h.handle.paths = append(h.handle.paths, &pathInfo{
		paths:       router.Paths,
		methods:     []string{http.MethodGet},
		middlewares: middlewares,
		handle: func(ctx *Context) {
			router.Handler(ctx.Writer, ctx.Request)
		},
		pos:       pos,
		isSwagger: true,
	})
}

func (h *handlerServer) handlePath(path *pathInfo) {
	var handleFunc HandleFunc
	if path.inFs != nil {
		handleFunc = h.handleStaticFS(path)
	} else {
		handleFunc = h.handleRouter(path)
	}
	for _, method := range path.methods {
		root := h.trees.get(method)
		if root == nil {
			root = &node{}
			h.trees = append(h.trees, methodTree{
				method: method,
				root:   root,
			})
		}
		for _, p := range path.paths {
			maxParams := countParams(p)
			if h.maxParams < maxParams {
				h.maxParams = maxParams
			}
			maxSkippedNodes := countSlash(p)
			if h.maxSkippedNodes < maxSkippedNodes {
				h.maxSkippedNodes = maxSkippedNodes
			}
			err := root.addRoute(p, handleFunc)
			if err != nil {
				log.Fatal(fmt.Errorf("%v, pos: %v", err, path.pos))
			}
		}
	}
}

func (h *handlerServer) handleStaticPath(path string) string {
	end := len(path) - 1
	for ; end >= 0 && path[end] != '/'; end-- {
	}
	if end == -1 {
		return ""
	}
	return path[:end]
}

func (h *handlerServer) handleStaticFS(path *pathInfo) HandleFunc {
	pathS := h.handleStaticPath(path.paths[0])
	fileServer := http.StripPrefix(pathS, http.FileServer(path.inFs))
	return func(ctx *Context) {
		ctx.path = path
		ctx.Extensions = path.extensions
		ctx.ChildPath = path.childPath
		h.handleLogger(ctx)
		ctx.handlers = append(path.middlewares, func(ctx *Context) {
			if path.isFile {
				http.ServeFile(ctx.Writer, ctx.Request, fmt.Sprintf("%v", path.inFs))
				return
			}
			fileServer.ServeHTTP(ctx.Writer, ctx.Request)
		})
		ctx.Next()
	}
}

func (h *handlerServer) handleRouter(path *pathInfo) HandleFunc {
	// Pre-build handlers slices to avoid making +copy for each request
	if path.handlersWithExec == nil {
		path.handlersWithExec = make([]HandleFunc, len(path.middlewares)+1)
		copy(path.handlersWithExec, path.middlewares)
		path.handlersWithExec[len(path.middlewares)] = func(ctx *Context) {
			h.execRouter(ctx)
		}
	}
	return func(ctx *Context) {
		ctx.path = path
		ctx.Extensions = path.extensions
		ctx.ChildPath = path.childPath
		ctx.RouterSummary = path.summary
		h.handleLogger(ctx)
		ctx.handlers = path.handlersWithExec
		ctx.Next()
	}
}

func (h *handlerServer) handleError(ctx *Context, err error) {
	var errorFunc func(err error) any
	if h.handle.errorMap[ctx.ChildPath] != nil {
		errorFunc = h.handle.errorMap[ctx.ChildPath].errorFunc
	}
	if errorFunc == nil {
		return
	}
	resp := errorFunc(err)
	h.handleResponse(ctx, resp)
}

func (h *handlerServer) execRouter(ctx *Context) {
	path := ctx.path
	if path.handle != nil {
		path.handle(ctx)
		return
	}
	var err error
	var inputs []reflect.Value
	lastInputIdx := 0
	var ctxVal reflect.Value
	if len(path.inTypes) == 2 {
		inputs = make([]reflect.Value, 2)
		ctxVal = reflect.ValueOf(ctx)
		inputs[0] = ctxVal
		lastInputIdx = 1
	} else {
		inputs = make([]reflect.Value, 1)
		lastInputIdx = 0
	}
	if path.existsCtx && !ctxVal.IsValid() {
		ctxVal = reflect.ValueOf(ctx)
	}
	inputs[lastInputIdx], err = h.handleInParamToValue(ctx, ctxVal, path.inTypes[lastInputIdx], path.inParams)
	if err != nil {
		h.handleError(ctx, getHTTPError(err, validErrorCode))
		return
	}
	// internal jump judgment of the security
	if ctx.isRedirect {
		return
	}
	rs := path.value.Call(inputs)
	// internal jump judgment of the execution method
	if ctx.isRedirect {
		return
	}
	if len(rs) == 0 {
		return
	}
	if len(rs) > 1 {
		respErr, _ := rs[1].Interface().(error)
		if respErr != nil {
			h.handleError(ctx, respErr)
			return
		}
	}
	h.handleResponse(ctx, rs[0].Interface())
}

func (h *handlerServer) handleResponse(ctx *Context, resp any) {
	var contentType string
	var addContentType bool
	if fn, ok := resp.(ResponseHeader); ok {
		header := fn.GetHeader()
		for key, vals := range header {
			for _, val := range vals {
				ctx.Writer.Header().Add(key, val)
			}
		}
		contentType = header.Get("Content-Type")
	}
	var mediaType MediaType
	if contentType == "" {
		mediaType = h.getResponseMediaType(ctx)
		addContentType = true
	} else {
		mediaType = MediaType(contentType).MediaType()
	}
	status := 0
	if fn, ok := resp.(ResponseStatus); ok {
		status = fn.GetStatus()
	}
	if fn, ok := resp.(ResponseBody); ok {
		resp = fn.GetBody()
	}
	if r, ok := getFnByCovertInterface[io.ReadCloser](resp); ok {
		if addContentType {
			ctx.Writer.Header().Add("Content-Type", string(mediaType))
		}
		if status != 0 {
			ctx.Writer.WriteHeader(status)
		}
		if err := h.copyReader(ctx.Writer, r); err != nil && ctx.Logger() != nil {
			ctx.Logger().Error("copy response body failed: %v", err)
		}
		return
	}
	var body []byte
	var err error
	if body, err = mediaType.Marshaler(resp); err != nil {
		h.handleError(ctx, NewHTTPError(validErrorCode, err.Error()))
		return
	}
	if addContentType {
		ctx.Writer.Header().Add("Content-Type", string(mediaType))
	}
	if status != 0 {
		ctx.Writer.WriteHeader(status)
	}
	_, _ = ctx.Writer.Write(body)
}

func (h *handlerServer) copyReader(w ResponseWriter, r io.ReadCloser) error {
	defer r.Close()
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			w.Flush()
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

func (h *handlerServer) handleInParamToValue(ctx *Context, ctxVal reflect.Value, inType reflect.Type, ins []*inParam) (value reflect.Value, err error) {
	value = reflect.New(inType).Elem()
	for value.Kind() == reflect.Ptr {
		initPtr(value)
		value = value.Elem()
	}
	var formType MediaType
	for _, field := range ins {
		if field.inType == inTypeFile {
			formType = formMultipart
		} else if field.inType == inTypeForm && formType != formMultipart {
			formType = formUrlencoded
		}
	}
	switch formType {
	case formUrlencoded:
		if err = ctx.Request.ParseForm(); err != nil {
			return
		}
	case formMultipart:
		if err = ctx.Request.ParseMultipartForm(32 << 20); err != nil {
			return
		}
	}
	for _, in := range ins {
		inValue := h.getParamValue(value, in.deeps)
		switch in.inType {
		case inTypePath:
			if val, ok := ctx.Params.Get(in.values[0].name); ok {
				if err = h.handleParamByString(ctx, inValue, in.field, val); err != nil {
					return
				}
			}
		case inTypeQuery:
			err = h.handleParamByStringSlice(ctx, inValue, in.field, ctx.Query()[in.values[0].name])
			if err != nil {
				return
			}
		case inTypeHeader:
			if err = h.handleParamByString(ctx, inValue, in.field, ctx.Request.Header.Get(in.values[0].name)); err != nil {
				return
			}
		case inTypeCookie:
			cookie, _ := ctx.Request.Cookie(in.values[0].name)
			if in.field._type.ConvertibleTo(typeCookie) {
				if err = h.handleParamByCookie(ctx, inValue, in.field, cookie); err != nil {
					return
				}
				continue
			}
			val := ""
			if cookie != nil {
				val = cookie.Value
			}
			if err = h.handleParamByString(ctx, inValue, in.field, val); err != nil {
				return
			}
		case inTypeForm:
			val := ""
			switch formType {
			case formUrlencoded:
				val = ctx.Request.Form.Get(in.values[0].name)
			case formMultipart:
				if ctx.Request.MultipartForm != nil && ctx.Request.MultipartForm.Value[in.values[0].name] != nil {
					val = ctx.Request.MultipartForm.Value[in.values[0].name][0]
				}
			}
			if err = h.handleParamByString(ctx, inValue, in.field, val); err != nil {
				return
			}
		case inTypeFile:
			var files []*multipart.FileHeader
			if ctx.Request.MultipartForm != nil {
				files = ctx.Request.MultipartForm.File[in.values[0].name]
			}
			if err = h.handleParamByFields(ctx, inValue, in.field, files); err != nil {
				return
			}
		case inTypeBody:
			mediaType := h.getRequestMediaType(ctx)
			if !h.isBodyMediaTypeAllowed(mediaType, in.values) {
				err = NewHTTPError(http.StatusUnsupportedMediaType, http.StatusText(http.StatusUnsupportedMediaType))
				return
			}
			err = h.setBody(inValue, ctx.Request.Body, mediaType)
			if err != nil {
				return
			}
			if mediaType.IsStream() {
				continue
			}
			err = h.validParamField(ctx, inValue, in.field, mediaType)
			if err != nil {
				return
			}
		case inTypeSecurityHTTPBearer:
			initPtr(inValue)
			authorization := ctx.Request.Header.Get("Authorization")
			authType, token, _ := strings.Cut(authorization, " ")
			if toFirstUpper(authType) != "Bearer" {
				token = ""
			}
			inValueAny := inValue.Interface()
			valOmitempty := false
			if fn, ok := inValueAny.(SecurityOmitempty); ok {
				valOmitempty = fn.Omitempty()
			}
			if !valOmitempty && token == "" {
				err = NewHTTPError(authErrorCode, ctx.lang().NotAuthenticated())
				return
			}
			security := inValueAny.(HTTPBearer)
			if err = security.HTTPBearer(token); err != nil {
				return
			}
		case inTypeSecurityHTTPBearerJWT:
			initPtr(inValue)
			authorization := ctx.Request.Header.Get("Authorization")
			authType, token, _ := strings.Cut(authorization, " ")
			if toFirstUpper(authType) != "Bearer" {
				token = ""
			}
			inValueAny := inValue.Interface()
			security := inValueAny.(HTTPBearerJWT)
			valOmitempty := false
			if fn, ok := inValueAny.(SecurityOmitempty); ok {
				valOmitempty = fn.Omitempty()
			}
			if token == "" {
				if valOmitempty {
					if err = security.HTTPBearerJWT(nil); err != nil {
						return
					}
					continue
				}
				err = NewHTTPError(authErrorCode, ctx.lang().NotAuthenticated())
				return
			}
			jwt := &JWT{}
			if err = decryptJWT(jwt, token, security); err != nil {
				err = NewHTTPError(authErrorCode, ctx.lang().JwtTranslate(err.Error()))
				return
			}
			if err = security.HTTPBearerJWT(jwt); err != nil {
				return
			}
		case inTypeSecurityHTTPBasic:
			initPtr(inValue)
			username, password, _ := ctx.Request.BasicAuth()
			inValueAny := inValue.Interface()
			valOmitempty := false
			if fn, ok := inValueAny.(SecurityOmitempty); ok {
				valOmitempty = fn.Omitempty()
			}
			if !valOmitempty && username == "" {
				err = NewHTTPError(authErrorCode, ctx.lang().NotAuthenticated())
				return
			}
			security := inValueAny.(HTTPBasic)
			if err = security.HTTPBasic(username, password); err != nil {
				return
			}
		case inTypeSecurityApiKey:
			initPtr(inValue)
			security := inValue.Interface().(ApiKey)
			if err = security.ApiKey(); err != nil {
				return
			}
		case inTypeCtx:
			h.handleParamByCtx(ctxVal, inValue)
		}
	}
	return
}

func (h *handlerServer) validParamField(ctx *Context, value reflect.Value, field *paramField, mediaType MediaType) (err error) {
	name := field.names.getFieldName(mediaType)
	desc := h.getDesc(name.name, field)
	if !field.anonymous {
		if value.Kind() != reflect.Ptr {
			if value.IsZero() {
				if name.required {
					return errors.New(ctx.lang().Required(desc))
				}
				if defaultSet(value, field.meta._default) {
					return h.validParamField(ctx, value, field, mediaType)
				}
				return
			}
		} else {
			for value.Kind() == reflect.Ptr {
				if value.IsNil() {
					if name.required {
						return errors.New(ctx.lang().Required(desc))
					}
					if defaultSet(value, field.meta._default) {
						return h.validParamField(ctx, value, field, mediaType)
					}
					return
				}
				if _, ok := getTypeByCovertInterface[TextInterface](value); ok {
					break
				}
				value = value.Elem()
			}
		}
	}
	realValue := value
	for value.Kind() == reflect.Ptr {
		initPtr(value)
		realValue = value
		value = value.Elem()
	}
	if err = h.handleValidate(field, realValue); err != nil {
		return
	}
	switch field.kind {
	case reflect.Struct:
		fields := field.fields
		if field.pkgName != "" {
			if sInfo := h.handle.structs[field.pkgName]; sInfo != nil {
				fields = sInfo.fields
			}
		}
		for _, childField := range fields {
			if err = h.validParamField(ctx, value.Field(childField.index), childField, mediaType); err != nil {
				return
			}
		}
	case reflect.Slice, reflect.Array:
		if field.meta.max != nil && uint64(value.Len()) > *field.meta.max {
			return errors.New(ctx.lang().Max(desc, *field.meta.max))
		}
		if uint64(value.Len()) < field.meta.min {
			return errors.New(ctx.lang().Min(desc, field.meta.min))
		}
		if field.meta.unique {
			m := map[any]struct{}{}
			for i := 0; i < value.Len(); i++ {
				itemVal := value.Index(i)
				if !itemVal.Comparable() {
					continue
				}
				item := itemVal.Interface()
				if _, ok := m[item]; ok {
					return errors.New(ctx.lang().Unique(desc))
				}
				m[item] = struct{}{}
			}
		}
		for i := 0; i < value.Len(); i++ {
			if err = h.validParamField(ctx, value.Index(i), field.fields[0], mediaType); err != nil {
				return
			}
		}
	case reflect.Map:
		if field.meta.max != nil && uint64(value.Len()) > *field.meta.max {
			return errors.New(ctx.lang().Max(desc, *field.meta.max))
		}
		if uint64(value.Len()) < field.meta.min {
			return errors.New(ctx.lang().Min(desc, field.meta.min))
		}
		for _, key := range value.MapKeys() {
			if err = h.validParamField(ctx, key, field.fields[0], mediaType); err != nil {
				return
			}
			if err = h.validParamField(ctx, value.MapIndex(key), field.fields[1], mediaType); err != nil {
				return
			}
		}
	case reflect.String:
		valStr := ""
		var enum []any
		if field.meta.enum != nil {
			enum = make([]any, len(field.meta.enum))
			copy(enum, field.meta.enum)
		}
		if field.isTextType {
			if fn, ok := getFnByCovertInterface[encoding.TextMarshaler](value); ok {
				var txt []byte
				if txt, err = fn.MarshalText(); err == nil {
					valStr = string(txt)
				}
			}
			for k, v := range enum {
				if fn, ok := getFnByCovertInterface[encoding.TextMarshaler](v); ok {
					var txt []byte
					if txt, err = fn.MarshalText(); err == nil {
						enum[k] = string(txt)
					}
				}
			}
		} else {
			valStr = value.String()
		}
		if field.meta.max != nil && uint64(len(valStr)) > *field.meta.max {
			return errors.New(ctx.lang().Max(desc, *field.meta.max))
		}
		if uint64(len(valStr)) < field.meta.min {
			return errors.New(ctx.lang().Min(desc, field.meta.min))
		}
		if field.meta.regexp != "" {
			if re := h.getCompiledRegexp(field.meta.regexp); re != nil && !re.MatchString(valStr) {
				return errors.New(ctx.lang().Regexp(desc, field.meta.regexp))
			}
		}
		if enum != nil && !inArrayAny(any(valStr), enum) {
			return errors.New(ctx.lang().Enum(desc, enum))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vFloat := float64(value.Int())
		if err = h.validFloat64(ctx, vFloat, desc, field); err != nil {
			return
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vFloat := float64(value.Uint())
		if err = h.validFloat64(ctx, vFloat, desc, field); err != nil {
			return
		}
	case reflect.Float32, reflect.Float64:
		vFloat := value.Float()
		if err = h.validFloat64(ctx, vFloat, desc, field); err != nil {
			return
		}
	default:
	}
	return
}

// getCompiledRegexp Return the compiled regular expression and use caching to avoid repeated compilation for each request
func (h *handlerServer) getCompiledRegexp(expr string) *regexp.Regexp {
	if expr == "" {
		return nil
	}
	if v, ok := h.regexpCache.Load(expr); ok {
		return v.(*regexp.Regexp)
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil
	}
	h.regexpCache.Store(expr, re)
	return re
}

func (h *handlerServer) getParamStringSlice(fType reflect.Type, value string) (values []string) {
	if value != "" {
		for fType.Kind() == reflect.Ptr {
			fType = fType.Elem()
		}
		values = []string{value}
		if fType.Kind() == reflect.Slice || fType.Kind() == reflect.Array {
			values = []string{}
			valList := strings.Split(value, ",")
			for _, v := range valList {
				values = append(values, strings.TrimSpace(v))
			}
		}
	}
	return
}

func (h *handlerServer) handleParamByFields(ctx *Context, value reflect.Value, field *paramField, fields []*multipart.FileHeader) (err error) {
	name := field.names.getFieldName("")
	desc := h.getDesc(name.name, field)
	if len(fields) == 0 || fields[0] == nil {
		if name.required {
			return errors.New(ctx.lang().Required(desc))
		}
		return
	}
	realValue := value
	for value.Kind() == reflect.Ptr {
		if value.Type().ConvertibleTo(typeFile) {
			valueSet(value, reflect.ValueOf(fields[0]))
			if err = h.handleValidate(field, value); err != nil {
				return
			}
			return
		}
		initPtr(value)
		realValue = value
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		newValue := reflect.MakeSlice(value.Type(), len(fields), len(fields))
		for i := 0; i < len(fields); i++ {
			if err = h.handleParamByFields(ctx, newValue.Index(i), field.fields[0], []*multipart.FileHeader{fields[i]}); err != nil {
				return
			}
		}
		value.Set(newValue)
		if err = h.handleValidate(field, realValue); err != nil {
			return
		}
	default:
	}
	return
}

func (h *handlerServer) handleParamByCookie(ctx *Context, value reflect.Value, field *paramField, cookie *http.Cookie) (err error) {
	name := field.names.getFieldName("")
	desc := h.getDesc(name.name, field)
	if cookie == nil || cookie.Value == "" {
		if name.required {
			return errors.New(ctx.lang().Required(desc))
		}
		return
	}
	for value.Kind() == reflect.Ptr {
		if value.Type().ConvertibleTo(typeCookie) {
			valueSet(value, reflect.ValueOf(cookie))
			if err = h.handleValidate(field, value); err != nil {
				return
			}
			return
		}
		initPtr(value)
		value = value.Elem()
	}
	return
}

func (h *handlerServer) handleParamByString(ctx *Context, value reflect.Value, field *paramField, val string) (err error) {
	name := field.names.getFieldName("")
	desc := h.getDesc(name.name, field)
	if val == "" {
		if name.required {
			return errors.New(ctx.lang().Required(desc))
		}
		if field.meta._defaultParamString != "" {
			return h.handleParamByString(ctx, value, field, field.meta._defaultParamString)
		}
		return
	}
	realValue := value
	for value.Kind() == reflect.Ptr {
		if value.Type().ConvertibleTo(typeFile) || value.Type().ConvertibleTo(typeCookie) {
			break
		}
		initPtr(value)
		realValue = value
		if _, ok := getTypeByCovertInterface[TextInterface](value); ok {
			break
		}
		value = value.Elem()
	}
	switch field.kind {
	case reflect.Slice, reflect.Array:
		values := h.getParamStringSlice(field._type, val)
		if err = h.handleParamByStringSlice(ctx, realValue, field, values); err != nil {
			return
		}
	case reflect.String:
		if field.meta.max != nil && uint64(len(val)) > *field.meta.max {
			return errors.New(ctx.lang().Max(desc, *field.meta.max))
		}
		if uint64(len(val)) < field.meta.min {
			return errors.New(ctx.lang().Min(desc, field.meta.min))
		}
		if field.meta.regexp != "" {
			if re := h.getCompiledRegexp(field.meta.regexp); re != nil && !re.MatchString(val) {
				return errors.New(ctx.lang().Regexp(desc, field.meta.regexp))
			}
		}
		var enum []any
		if field.meta.enum != nil {
			enum = make([]any, len(field.meta.enum))
			copy(enum, field.meta.enum)
		}
		if field.isTextType {
			for k, v := range enum {
				if fn, ok := getFnByCovertInterface[encoding.TextMarshaler](v); ok {
					var txt []byte
					if txt, err = fn.MarshalText(); err == nil {
						enum[k] = string(txt)
					}
				}
			}
		}
		if enum != nil && !inArrayAny(any(val), enum) {
			return errors.New(ctx.lang().Enum(desc, field.meta.enum))
		}
		if field.isTextType {
			if err = coverInterfaceByValue[TextInterface](value, func(fn TextInterface) error {
				return fn.UnmarshalText([]byte(val))
			}, true); err != nil {
				return
			}
		} else {
			value.SetString(val)
		}
		if err = h.handleValidate(field, realValue); err != nil {
			return
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var valInt int64
		valInt, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return
		}
		// When using required keys and zero values, please use Pointers
		if valInt == 0 {
			if name.required {
				return errors.New(ctx.lang().Required(desc))
			}
			return
		}
		if err = h.validFloat64(ctx, float64(valInt), desc, field); err != nil {
			return
		}
		value.SetInt(valInt)
		if err = h.handleValidate(field, realValue); err != nil {
			return
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var valUint uint64
		valUint, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			return
		}
		// When using required keys and zero values, please use Pointers
		if valUint == 0 {
			if name.required {
				return errors.New(ctx.lang().Required(desc))
			}
			return
		}
		if err = h.validFloat64(ctx, float64(valUint), desc, field); err != nil {
			return
		}
		value.SetUint(valUint)
		if err = h.handleValidate(field, realValue); err != nil {
			return
		}
	case reflect.Float32, reflect.Float64:
		var valFloat float64
		valFloat, err = strconv.ParseFloat(val, 64)
		if err != nil {
			return
		}
		// When using required keys and zero values, please use Pointers
		if valFloat == 0 {
			if name.required {
				return errors.New(ctx.lang().Required(desc))
			}
			return
		}
		if err = h.validFloat64(ctx, valFloat, desc, field); err != nil {
			return
		}
		value.SetFloat(valFloat)
		if err = h.handleValidate(field, realValue); err != nil {
			return
		}
	case reflect.Bool:
		var valBool bool
		valBool, err = strconv.ParseBool(val)
		if err != nil {
			return
		}
		if field.meta.enum != nil && !inArrayAny(any(valBool), field.meta.enum) {
			return errors.New(ctx.lang().Enum(desc, field.meta.enum))
		}
		value.SetBool(valBool)
		if err = h.handleValidate(field, realValue); err != nil {
			return
		}
	default:
	}
	return
}

func (h *handlerServer) handleParamByStringSlice(ctx *Context, value reflect.Value, field *paramField, values []string) (err error) {
	name := field.names.getFieldName("")
	desc := h.getDesc(name.name, field)
	if len(values) == 0 || values[0] == "" {
		if name.required {
			return errors.New(ctx.lang().Required(desc))
		}
		if field.meta._defaultParamString != "" {
			return h.handleParamByString(ctx, value, field, field.meta._defaultParamString)
		}
		return
	}
	realValue := value
	for value.Kind() == reflect.Ptr {
		if value.Type().ConvertibleTo(typeFile) || value.Type().ConvertibleTo(typeCookie) {
			break
		}
		initPtr(value)
		realValue = value
		if _, ok := getTypeByCovertInterface[TextInterface](value); ok {
			break
		}
		value = value.Elem()
	}
	switch field.kind {
	case reflect.Slice, reflect.Array:
		if field.meta.max != nil && uint64(len(values)) > *field.meta.max {
			return errors.New(ctx.lang().Max(desc, *field.meta.max))
		}
		if uint64(len(values)) < field.meta.min {
			return errors.New(ctx.lang().Min(desc, field.meta.min))
		}
		if field.meta.unique {
			m := map[any]struct{}{}
			for _, val := range values {
				if _, ok := m[val]; ok {
					return errors.New(ctx.lang().Unique(desc))
				}
				m[val] = struct{}{}
			}
		}
		var newValue reflect.Value
		if field.kind == reflect.Slice {
			newValue = reflect.MakeSlice(value.Type(), len(values), len(values))
		} else {
			newValue = reflect.New(value.Type()).Elem()
		}
		valLen := newValue.Len()
		for i := 0; i < len(values); i++ {
			if i < valLen {
				childVal := newValue.Index(i)
				if err = h.handleParamByStringSlice(ctx, childVal, field.fields[0], []string{values[i]}); err != nil {
					return
				}
			}
		}
		valueSet(value, newValue)
		if err = h.handleValidate(field, realValue); err != nil {
			return
		}
	default:
		if err = h.handleParamByString(ctx, realValue, field, values[0]); err != nil {
			return
		}
	}
	return
}

func (h *handlerServer) handleValidate(field *paramField, value reflect.Value) error {
	if !field.meta.isValid {
		return nil
	}
	if fn, ok := getFnByCovertInterface[MetaValidate](value); ok {
		return fn.Validate()
	}
	return nil
}

func (h *handlerServer) handleParamByCtx(ctxVal reflect.Value, value reflect.Value) {
	for value.Kind() == reflect.Ptr {
		if value.Type() == typeContext {
			value.Set(ctxVal)
			return
		}
		initPtr(value)
		value = value.Elem()
	}
}

func (h *handlerServer) validFloat64(ctx *Context, vFloat float64, desc string, field *paramField) (err error) {
	if field.meta.lt != nil && vFloat >= *field.meta.lt {
		return errors.New(ctx.lang().Lt(desc, *field.meta.lt))
	}
	if field.meta.lte != nil && vFloat > *field.meta.lte {
		return errors.New(ctx.lang().Lte(desc, *field.meta.lte))
	}
	if field.meta.gt != nil && vFloat <= *field.meta.gt {
		return errors.New(ctx.lang().Gt(desc, *field.meta.gt))
	}
	if field.meta.gte != nil && vFloat < *field.meta.gte {
		return errors.New(ctx.lang().Gte(desc, *field.meta.gte))
	}
	if field.meta.multiple != nil {
		if *field.meta.multiple == 0 {
			return errors.New(ctx.lang().MultipleOf(desc, *field.meta.multiple))
		}
		rs, _ := decimal.NewFromFloat(vFloat).Div(decimal.NewFromFloat(*field.meta.multiple)).Float64()
		if rs != float64(int64(rs)) {
			return errors.New(ctx.lang().MultipleOf(desc, *field.meta.multiple))
		}
	}
	if field.meta.enum != nil && !inArrayAny(any(vFloat), field.meta.enum) {
		return errors.New(ctx.lang().Enum(desc, field.meta.enum))
	}
	return
}

func (h *handlerServer) removeMorPtrValue(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Ptr {
		initPtr(value)
		if value.Elem().Kind() != reflect.Ptr {
			return value
		}
		value = value.Elem()
	}
	return value
}

func (h *handlerServer) getParamValue(value reflect.Value, deeps []int) reflect.Value {
	for len(deeps) > 0 {
		index := deeps[0]
		deeps = deeps[1:]
		for value.Kind() == reflect.Ptr {
			initPtr(value)
			value = value.Elem()
		}
		value = value.Field(index)
	}
	return value
}

func (h *handlerServer) setBody(value reflect.Value, reader io.ReadCloser, mediaType MediaType) (err error) {
	if reader == nil {
		return nil
	}
	value = h.removeMorPtrValue(value)
	if value.Kind() == reflect.Ptr {
		return mediaType.Unmarshaler(reader, value)
	}
	newValue := reflect.New(value.Type())
	if err = mediaType.Unmarshaler(reader, newValue); err != nil {
		return
	}
	value.Set(newValue.Elem())
	return
}

func (h *handlerServer) isBodyMediaTypeAllowed(mediaType MediaType, allowed paramFieldNames) bool {
	mediaType = mediaType.MediaType()
	for _, item := range allowed {
		if mediaType == item.mediaType.MediaType() {
			return true
		}
	}
	return false
}

func (h *handlerServer) getMiddlewares(path string) (rs []HandleFunc) {
	pathList := strings.Split(path, "/")
	match := ""
	for _, val := range pathList {
		match += "/" + val
		groupMiddles := h.handle.publicGroupMiddlewares[match[1:]]
		if groupMiddles == nil {
			continue
		}
		rs = append(rs, groupMiddles...)
	}
	return
}

func (h *handlerServer) getChildPath(path string) (childPath string) {
	var ok bool
	for {
		if _, ok = h.handle.childMap[path]; ok {
			childPath = path
			return
		}
		index := strings.LastIndex(path, "/")
		if index == -1 {
			return
		}
		path = path[:index]
	}
}

func (h *handlerServer) notFind(ctx *Context) {
	ctx.handlers = h.getMiddlewares(ctx.Request.URL.Path)
	ctx.handlers = append(ctx.handlers, func(ctx *Context) {
		child := h.handle.childMap[ctx.ChildPath]
		if child.handleMethodNotAllowed {
			allowed := make([]string, 0, len(h.trees)-1)
			for _, tree := range h.trees {
				if tree.method == ctx.Request.Method {
					continue
				}
				params := make(Params, 0, h.maxParams)
				skippedNodes := make([]skippedNode, 0, h.maxSkippedNodes)
				if val := tree.root.getValue(ctx.Request.URL.Path, &params, &skippedNodes); val.handler != nil {
					allowed = append(allowed, tree.method)
				}
			}
			if len(allowed) > 0 {
				ctx.Writer.Header().Set("Allow", strings.Join(allowed, ", "))
				child.noMethod(ctx)
				return
			}
		}
		child.noRoute(ctx)
	})
	ctx.Next()
}

func (h *handlerServer) redirect(ctx *Context) {
	ctx.handlers = h.getMiddlewares(ctx.Request.URL.Path)
	ctx.handlers = append(ctx.handlers, func(ctx *Context) {
		tsrPath := h.handleTsrPath(ctx.Request.URL.Path)
		if ctx.Request.URL.RawQuery != "" {
			tsrPath += "?" + ctx.Request.URL.RawQuery
		}
		code := http.StatusMovedPermanently
		if ctx.Request.Method != http.MethodGet {
			code = http.StatusTemporaryRedirect
		}
		ctx.Redirect(code, tsrPath)
	})
	ctx.Next()
}

func (h *handlerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.pool.Get().(*Context)
	ctx.baseLog = h.log
	ctx.log = h.log
	ctx.writermem.reset(w)
	ctx.reset()
	ctx.Request = r
	if ctx.handleError == nil {
		ctx.handleError = h.handleError
	}
	h.generateRequestID(ctx)
	ctx.langInfo = h.parseAcceptLanguage(ctx, h.handle.langList)
	h.handleHTTPRequest(ctx)
	h.pool.Put(ctx)
}

func (h *handlerServer) generateRequestID(ctx *Context) {
	if h.handle.api.GenerateRequestID {
		if h.handle.api.UseXRequestIDHeader && ctx.Request != nil {
			if requestID := strings.TrimSpace(ctx.Request.Header.Get("X-Request-ID")); requestID != "" {
				ctx.RequestID = requestID
				h.setXRequestIDHeader(ctx)
				return
			}
		}
		pk, _ := uuid.NewV4()
		ctx.RequestID = pk.String()
		h.setXRequestIDHeader(ctx)
	}
}

func (h *handlerServer) setXRequestIDHeader(ctx *Context) {
	if !h.handle.api.UseXRequestIDHeader || ctx.Writer == nil || ctx.RequestID == "" {
		return
	}
	ctx.Writer.Header().Set("X-Request-ID", ctx.RequestID)
}

func (h *handlerServer) handleLogger(ctx *Context) {
	baseLog := ctx.baseLog
	if baseLog == nil {
		baseLog = ctx.log
	}
	if baseLog == nil {
		return
	}
	ctx.log = baseLog
	if fn, ok := getFnByCovertInterface[LoggerWithContext](baseLog); ok {
		if logger := fn.WithContext(ctx); logger != nil {
			ctx.log = logger
		}
	}
}

func (h *handlerServer) handleHTTPRequest(ctx *Context) {
	root := h.trees.get(ctx.Request.Method)
	if root == nil {
		ctx.ChildPath = h.getChildPath(ctx.Request.URL.Path)
		h.handleLogger(ctx)
		h.notFind(ctx)
		return
	}
	value := root.getValue(ctx.Request.URL.Path, ctx.Params, ctx.skippedNodes)
	if value.handler != nil {
		ctx.fullPath = value.fullPath
		value.handler(ctx)
		return
	}
	*ctx.Params = (*ctx.Params)[:0]
	*ctx.skippedNodes = (*ctx.skippedNodes)[:0]
	ctx.ChildPath = h.getChildPath(ctx.Request.URL.Path)
	h.handleLogger(ctx)
	if value.tsr {
		child := h.handle.childMap[ctx.ChildPath]
		if child.redirectTrailingSlash {
			h.redirect(ctx)
			return
		}
	}
	h.notFind(ctx)
}

func (h *handlerServer) handleTsrPath(path string) string {
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	} else {
		path = path + "/"
	}
	return path
}

func (h *handlerServer) getRequestMediaType(ctx *Context) MediaType {
	contentType, _, _ := strings.Cut(ctx.Request.Header.Get("Content-Type"), ";")
	return MediaType(contentType)
}

func (h *handlerServer) getResponseMediaType(ctx *Context) MediaType {
	child := h.handle.childMap[ctx.ChildPath]
	if len(child.responseMediaTypes) == 1 {
		return child.responseMediaTypes[0]
	}
	if child.useMediaType {
		mediaTypeStr := ctx.Request.URL.Query().Get(returnMediaTypeField)
		if mediaTypeStr != "" {
			mediaType := MediaType(mediaTypeStr).MediaType()
			if inArray(mediaType, child.responseMediaTypes) {
				return mediaType
			}
		}
	}
	return h.parseAccept(ctx.Request.Header.Get("Accept"), child.responseMediaTypes)
}

func (h *handlerServer) parseAccept(accept string, responseMediaTypes []MediaType) (mediaType MediaType) {
	parts := strings.Split(accept, ",")
	var mediaTypeQ float64 = -1
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		mediaTypeStr, params, err := mime.ParseMediaType(part)
		if err != nil {
			continue
		}
		q := 1.0
		if qStr, ok := params["q"]; ok {
			qVal, qErr := strconv.ParseFloat(qStr, 64)
			if qErr == nil {
				q = qVal
				if q < 0 {
					q = 0
				}
				if q > 1 {
					q = 1
				}
			}
		}
		// no match
		if q == 0 {
			continue
		}
		if mediaTypeQ != -1 && mediaTypeQ >= q {
			continue
		}
		if mediaTypeStr == "*/*" {
			mediaTypeQ = q
			mediaType = responseMediaTypes[0]
		} else if strings.HasSuffix(mediaTypeStr, "/*") {
			prefix := mediaTypeStr[:len(mediaTypeStr)-1]
			for _, mt := range responseMediaTypes {
				if strings.HasPrefix(string(mt), prefix) {
					mediaTypeQ = q
					mediaType = mt
					break
				}
			}
		} else {
			mType := MediaType(mediaTypeStr)
			if inArray(mType, responseMediaTypes) {
				mediaTypeQ = q
				mediaType = mType
			}
		}
	}
	if mediaType == "" {
		return responseMediaTypes[0]
	}
	return
}

func (h *handlerServer) parseAcceptLanguage(ctx *Context, langList []Lang) (lang Lang) {
	if len(langList) == 1 {
		return langList[0]
	}
	acceptLanguage := ctx.Request.Header.Get("Accept-Language")
	acceptLanguage = strings.TrimSpace(acceptLanguage)
	if acceptLanguage == "" {
		return langList[0]
	}
	parts := strings.Split(acceptLanguage, ",")
	var langQ float64 = -1
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		langStr, paramsStr, ok := strings.Cut(part, ";")
		langStr = strings.TrimSpace(langStr)
		q := 1.0
		if ok {
			paramsStr = strings.TrimSpace(paramsStr)
			if paramsStr != "" {
				pList := strings.Split(paramsStr, ";")
				for _, v := range pList {
					qKey, qValStr, qOk := strings.Cut(v, "=")
					if !qOk {
						continue
					}
					if strings.TrimSpace(qKey) == "q" {
						qValStr = strings.TrimSpace(qValStr)
						qVal, qErr := strconv.ParseFloat(qValStr, 64)
						if qErr == nil {
							q = qVal
							if q < 0 {
								q = 0
							}
							if q > 1 {
								q = 1
							}
						}
						break
					}
				}
			}
		}
		// no match
		if q == 0 {
			continue
		}
		if langQ != -1 && langQ >= q {
			continue
		}
		if h.handle.langMap[langStr] != nil {
			langQ = q
			lang = h.handle.langMap[langStr]
			if q == 1.0 {
				return
			}
		}
	}
	if lang == nil {
		return langList[0]
	}
	return
}

func (h *handlerServer) getDesc(fieldName string, field *paramField) string {
	if fieldName == "" {
		fieldName = field.name
	}
	if field.meta.name != "" {
		return field.meta.name
	}
	if field.meta.desc != "" {
		return field.meta.desc
	}
	return fieldName
}
