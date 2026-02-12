package goapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"runtime/debug"
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
		return &Context{}
	}
	return hs
}

type handlerServer struct {
	log    Logger
	handle *handler
	trees  methodTrees
	pool   sync.Pool
}

func (h *handlerServer) Handle() {
	for _, path := range h.handle.paths {
		h.handlePath(path)
	}
	debugPrintRouter(h.log, h.handle.paths)
}

func (h *handlerServer) HandleSwagger(
	fn func(path, title string, openapiJsonBody []byte, config swagger.Config) (routers []swagger.Router),
	openapiMap map[string]*openapi.OpenAPI,
) {
	for docsPath, openAPI := range openapiMap {
		if err := openAPI.Validate(); err != nil {
			log.Fatal(err)
		}
		openapiBody, _ := json.Marshal(openAPI)
		routers := fn(docsPath, openAPI.Info.Title, openapiBody, h.handle.swaggerMap[docsPath])
		for _, router := range routers {
			h.handleSwagger(router)
		}
	}
}

func (h *handlerServer) handleSwagger(router swagger.Router) {
	pos := "github.com/goodluckxu-go/goapi/v2/swagger.GetSwagger (docs)"
	middlewares := h.getMiddlewares(router.Paths[0])
	if len(middlewares) > 0 {
		pos += fmt.Sprintf(" (%v Middleware)", len(middlewares))
	}
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
			err := root.addRoute(p, handleFunc)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (h *handlerServer) handleStaticPath(path string) string {
	end := len(path) - 1
	for ; end >= 0 && path[end] != '/'; end-- {
	}
	return path[:end]
}

func (h *handlerServer) handleStaticFS(path *pathInfo) HandleFunc {
	pathS := h.handleStaticPath(path.paths[0])
	fileServer := http.StripPrefix(pathS, http.FileServer(path.inFs))
	return func(ctx *Context) {
		ctx.path = path
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
	return func(ctx *Context) {
		ctx.path = path
		ctx.ChildPath = path.childPath
		h.handleLogger(ctx)
		ctx.handlers = make([]HandleFunc, len(path.middlewares)+1)
		n := copy(ctx.handlers, path.middlewares)
		ctx.handlers[n] = func(ctx *Context) {
			h.execRouter(ctx)
		}
		ctx.Next()
	}
}

func (h *handlerServer) handleExcept(ctx *Context, err string, code ...int) {
	var exceptFunc func(httpCode int, detail string) any
	if h.handle.exceptMap[ctx.ChildPath] != nil {
		exceptFunc = h.handle.exceptMap[ctx.ChildPath].exceptFunc
	}
	if exceptFunc == nil {
		return
	}
	if len(code) > 0 {
		resp := exceptFunc(code[0], err)
		h.handleResponse(ctx, resp)
		return
	}
	var res exceptJson
	er := json.Unmarshal([]byte(err), &res)
	var resp any
	if er != nil {
		resp = exceptFunc(http.StatusInternalServerError, err)
		h.handle.api.log.Error("panic: %v [recovered]\n%v", er, string(debug.Stack()))
	} else {
		resp = exceptFunc(res.HttpCode, res.Detail)
	}
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
	if len(path.inTypes) == 2 {
		inputs = make([]reflect.Value, 2)
		inputs[0] = reflect.ValueOf(ctx)
		lastInputIdx = 1
	} else {
		inputs = make([]reflect.Value, 1)
		lastInputIdx = 0
	}
	inputs[lastInputIdx], err = h.handleInParamToValue(ctx, path.inTypes[lastInputIdx], path.inParams)
	if err != nil {
		h.handleExcept(ctx, err.Error(), validErrorCode)
		return
	}
	rs := path.value.Call(inputs)
	if len(rs) == 0 {
		return
	}
	h.handleResponse(ctx, rs[0].Interface())
}

func (h *handlerServer) handleResponse(ctx *Context, resp any) {
	mediaType := h.getResponseMediaType(ctx)
	var header http.Header
	if fn, ok := resp.(ResponseHeader); ok {
		header = fn.GetHeader()
	}
	if header == nil {
		header = make(http.Header)
	}
	contentType := header.Get("Content-Type")
	if contentType == "" {
		contentType = string(mediaType)
		header.Set("Content-Type", contentType)
	}
	mediaType = MediaType(contentType)
	for key, vals := range header {
		for _, val := range vals {
			ctx.Writer.Header().Add(key, val)
		}
	}
	if fn, ok := resp.(ResponseStatus); ok {
		ctx.Writer.WriteHeader(fn.GetStatus())
	}
	if fn, ok := resp.(ResponseBody); ok {
		resp = fn.GetBody()
	}
	if r, ok := getFnByCovertInterface[io.ReadCloser](resp); ok {
		_ = h.copyReader(ctx.Writer, r)
		return
	}
	var body []byte
	var err error
	if body, err = mediaType.Marshaler(resp); err != nil {
		h.handleExcept(ctx, err.Error(), validErrorCode)
		return
	}
	_, _ = ctx.Writer.Write(body)
}

func (h *handlerServer) copyReader(w ResponseWriter, r io.ReadCloser) error {
	defer r.Close()
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			_, _ = w.Write(buf[:n])
			w.Flush()
		}
		if err != nil {
			return err
		}
	}
}

func (h *handlerServer) handleInParamToValue(ctx *Context, inType reflect.Type, ins []*inParam) (value reflect.Value, err error) {
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
				values := h.getParamStringSlice(in.field._type, val)
				if err = h.handleParamByStringSlice(inValue, in.field, values); err != nil {
					return
				}
			}
		case inTypeQuery:
			err = h.handleParamByStringSlice(inValue, in.field, ctx.Query()[in.values[0].name])
			if err != nil {
				return
			}
		case inTypeHeader:
			values := h.getParamStringSlice(in.field._type, ctx.Request.Header.Get(in.values[0].name))
			if err = h.handleParamByStringSlice(inValue, in.field, values); err != nil {
				return
			}
		case inTypeCookie:
			cookie, _ := ctx.Request.Cookie(in.values[0].name)
			if in.field._type.ConvertibleTo(typeCookie) {
				if err = h.handleParamByCookie(inValue, in.field, cookie); err != nil {
					return
				}
				continue
			}
			val := ""
			if cookie != nil {
				val = cookie.Value
			}
			values := h.getParamStringSlice(in.field._type, val)
			if err = h.handleParamByStringSlice(inValue, in.field, values); err != nil {
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
			values := h.getParamStringSlice(in.field._type, val)
			if err = h.handleParamByStringSlice(inValue, in.field, values); err != nil {
				return
			}
		case inTypeFile:
			var files []*multipart.FileHeader
			if ctx.Request.MultipartForm != nil {
				files = ctx.Request.MultipartForm.File[in.values[0].name]
			}
			if err = h.handleParamByFields(inValue, in.field, files); err != nil {
				return
			}
		case inTypeBody:
			mediaType := h.getRequestMediaType(ctx)
			if ctx.Request.ContentLength > 0 {
				err = h.setBody(inValue, ctx.Request.Body, mediaType)
				if err != nil {
					return
				}
			}
			if mediaType.IsStream() {
				continue
			}
			err = h.validParamField(inValue, in.field, mediaType)
			if err != nil {
				return
			}
		case inTypeSecurityHTTPBearer:
			initPtr(inValue)
			authorization := ctx.Request.Header.Get("Authorization")
			authList := strings.Split(authorization, " ")
			token := ""
			if len(authList) == 2 && authList[0] == "Bearer" {
				token = authList[1]
			}
			security := inValue.Interface().(HTTPBearer)
			security.HTTPBearer(token)
		case inTypeSecurityHTTPBearerJWT:
			initPtr(inValue)
			authorization := ctx.Request.Header.Get("Authorization")
			authList := strings.Split(authorization, " ")
			token := ""
			if len(authList) == 2 && authList[0] == "Bearer" {
				token = authList[1]
			}
			security := inValue.Interface().(HTTPBearerJWT)
			jwt := &JWT{}
			if err = decryptJWT(jwt, token, security); err != nil {
				HTTPException(authErrorCode, h.handle.api.lang.JwtTranslate(err.Error()))
			}
			security.HTTPBearerJWT(jwt)
		case inTypeSecurityHTTPBasic:
			initPtr(inValue)
			username, password, _ := ctx.Request.BasicAuth()
			security := inValue.Interface().(HTTPBasic)
			security.HTTPBasic(username, password)
		case inTypeSecurityApiKey:
			initPtr(inValue)
			security := inValue.Interface().(ApiKey)
			security.ApiKey()
		case inTypeOther:
			h.handleParamByOther(ctx, inValue)
		}
	}
	return
}

func (h *handlerServer) validParamField(value reflect.Value, field *paramField, mediaType MediaType) (err error) {
	name := field.names.getFieldName(mediaType)
	desc := name.name
	if desc == "" {
		desc = field.name
	}
	if field.tag.desc != "" {
		desc = field.tag.desc
	}
	if !field.anonymous {
		if value.Kind() != reflect.Ptr {
			if value.IsZero() {
				if name.required {
					return errors.New(h.handle.api.lang.Required(desc))
				}
				return
			}
		} else {
			for value.Kind() == reflect.Ptr {
				if value.IsNil() {
					if name.required {
						return errors.New(h.handle.api.lang.Required(desc))
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
	switch field.kind {
	case reflect.Struct:
		fields := field.fields
		if field.pkgName != "" {
			sInfo := h.handle.structs[field.pkgName]
			fields = sInfo.fields
		}
		for _, childField := range fields {
			if err = h.validParamField(value.Field(childField.index), childField, mediaType); err != nil {
				return
			}
		}
	case reflect.Slice, reflect.Array:
		if field.tag.max != nil && uint64(value.Len()) > *field.tag.max {
			return errors.New(h.handle.api.lang.Max(desc, *field.tag.max))
		}
		if uint64(value.Len()) < field.tag.min {
			return errors.New(h.handle.api.lang.Min(desc, field.tag.min))
		}
		if field.tag.unique {
			m := map[any]struct{}{}
			for i := 0; i < value.Len(); i++ {
				if _, ok := m[value.Index(i).Interface()]; ok {
					return errors.New(h.handle.api.lang.Unique(desc))
				}
				m[value.Index(i).Interface()] = struct{}{}
			}
		}
		for i := 0; i < value.Len(); i++ {
			if err = h.validParamField(value.Index(i), field.fields[0], mediaType); err != nil {
				return
			}
		}
	case reflect.Map:
		if field.tag.max != nil && uint64(value.Len()) > *field.tag.max {
			return errors.New(h.handle.api.lang.Max(desc, *field.tag.max))
		}
		if uint64(value.Len()) < field.tag.min {
			return errors.New(h.handle.api.lang.Min(desc, field.tag.min))
		}
		for _, key := range value.MapKeys() {
			if err = h.validParamField(key, field.fields[0], mediaType); err != nil {
				return
			}
			if err = h.validParamField(value.MapIndex(key), field.fields[1], mediaType); err != nil {
				return
			}
		}
	case reflect.String:
		valStr := ""
		if field.isTextType {
			if fn, ok := getFnByCovertInterface[TextInterface](value, true); ok {
				var txt []byte
				if txt, err = fn.MarshalText(); err == nil {
					valStr = string(txt)
				}
			}
		} else {
			valStr = value.String()
		}
		if field.tag.max != nil && uint64(len(valStr)) > *field.tag.max {
			return errors.New(h.handle.api.lang.Max(desc, *field.tag.max))
		}
		if uint64(len(valStr)) < field.tag.min {
			return errors.New(h.handle.api.lang.Min(desc, field.tag.min))
		}
		if field.tag.regexp != "" && !regexp.MustCompile(field.tag.regexp).MatchString(valStr) {
			return errors.New(h.handle.api.lang.Regexp(desc, field.tag.regexp))
		}
		if field.tag.enum != nil && !inArrayAny(any(valStr), field.tag.enum) {
			return errors.New(h.handle.api.lang.Enum(desc, field.tag.enum))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vFloat := float64(value.Int())
		if err = h.validFloat64(vFloat, desc, field); err != nil {
			return
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vFloat := float64(value.Uint())
		if err = h.validFloat64(vFloat, desc, field); err != nil {
			return
		}
	case reflect.Float32, reflect.Float64:
		vFloat := value.Float()
		if err = h.validFloat64(vFloat, desc, field); err != nil {
			return
		}
	default:
	}
	return
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

func (h *handlerServer) handleParamByFields(value reflect.Value, field *paramField, fields []*multipart.FileHeader) (err error) {
	name := field.names.getFieldName("")
	desc := name.name
	if field.tag.desc != "" {
		desc = field.tag.desc
	}
	if len(fields) == 0 || fields[0] == nil {
		if name.required {
			return errors.New(h.handle.api.lang.Required(desc))
		}
		return
	}
	for value.Kind() == reflect.Ptr {
		if value.Type().ConvertibleTo(typeFile) {
			value.Set(reflect.ValueOf(fields[0]).Convert(value.Type()))
			break
		}
		initPtr(value)
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		newValue := reflect.MakeSlice(value.Type(), len(fields), len(fields))
		for i := 0; i < len(fields); i++ {
			if err = h.handleParamByFields(newValue.Index(i), field.fields[0], []*multipart.FileHeader{fields[i]}); err != nil {
				return
			}
		}
		value.Set(newValue)
	default:
	}
	return
}

func (h *handlerServer) handleParamByCookie(value reflect.Value, field *paramField, cookie *http.Cookie) (err error) {
	name := field.names.getFieldName("")
	desc := name.name
	if field.tag.desc != "" {
		desc = field.tag.desc
	}
	if cookie == nil || cookie.Value == "" {
		if name.required {
			return errors.New(h.handle.api.lang.Required(desc))
		}
		return
	}
	for value.Kind() == reflect.Ptr {
		if value.Type().ConvertibleTo(typeCookie) {
			value.Set(reflect.ValueOf(cookie).Convert(value.Type()))
			return
		}
		initPtr(value)
		value = value.Elem()
	}
	return
}

func (h *handlerServer) handleParamByStringSlice(value reflect.Value, field *paramField, values []string) (err error) {
	name := field.names.getFieldName("")
	desc := name.name
	if field.tag.desc != "" {
		desc = field.tag.desc
	}
	if len(values) == 0 || values[0] == "" {
		if name.required {
			return errors.New(h.handle.api.lang.Required(desc))
		}
		return
	}
	for value.Kind() == reflect.Ptr {
		if value.Type().ConvertibleTo(typeFile) || value.Type().ConvertibleTo(typeCookie) {
			break
		}
		initPtr(value)
		if _, ok := getTypeByCovertInterface[TextInterface](value); ok {
			break
		}
		value = value.Elem()
	}
	switch field.kind {
	case reflect.Slice, reflect.Array:
		if field.tag.max != nil && uint64(len(values)) > *field.tag.max {
			return errors.New(h.handle.api.lang.Max(desc, *field.tag.max))
		}
		if uint64(len(values)) < field.tag.min {
			return errors.New(h.handle.api.lang.Min(desc, field.tag.min))
		}
		if field.tag.unique {
			m := map[any]struct{}{}
			for _, val := range values {
				if _, ok := m[val]; ok {
					return errors.New(h.handle.api.lang.Unique(desc))
				}
				m[val] = struct{}{}
			}
		}
		newValue := reflect.MakeSlice(value.Type(), len(values), len(values))
		for i := 0; i < len(values); i++ {
			childVal := newValue.Index(i)
			if err = h.handleParamByStringSlice(childVal, field.fields[0], []string{values[i]}); err != nil {
				return
			}
		}
		value.Set(newValue.Convert(value.Type()))
	case reflect.String:
		var valStr string
		if len(values) > 0 {
			valStr = values[0]
		}
		if field.tag.max != nil && uint64(len(valStr)) > *field.tag.max {
			return errors.New(h.handle.api.lang.Max(desc, *field.tag.max))
		}
		if uint64(len(valStr)) < field.tag.min {
			return errors.New(h.handle.api.lang.Min(desc, field.tag.min))
		}
		if field.tag.regexp != "" && !regexp.MustCompile(field.tag.regexp).MatchString(valStr) {
			return errors.New(h.handle.api.lang.Regexp(desc, field.tag.regexp))
		}
		if field.tag.enum != nil && !inArrayAny(any(valStr), field.tag.enum) {
			return errors.New(h.handle.api.lang.Enum(desc, field.tag.enum))
		}
		if field.isTextType {
			if err = coverInterfaceByValue[TextInterface](value, func(fn TextInterface) error {
				return fn.UnmarshalText([]byte(values[0]))
			}, true); err != nil {
				return
			}
		} else {
			value.Set(reflect.ValueOf(valStr).Convert(value.Type()))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var valInt int64
		if len(values) > 0 {
			valInt, err = strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return
			}
		}
		if valInt == 0 {
			if name.required {
				return errors.New(h.handle.api.lang.Required(desc))
			}
			return
		}
		if err = h.validFloat64(float64(valInt), desc, field); err != nil {
			return
		}
		value.Set(reflect.ValueOf(valInt).Convert(value.Type()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var valUint uint64
		if len(values) > 0 {
			valUint, err = strconv.ParseUint(values[0], 10, 64)
			if err != nil {
				return
			}
		}
		if valUint == 0 {
			if name.required {
				return errors.New(h.handle.api.lang.Required(desc))
			}
			return
		}
		if err = h.validFloat64(float64(valUint), desc, field); err != nil {
			return
		}
		value.Set(reflect.ValueOf(valUint).Convert(value.Type()))
	case reflect.Float32, reflect.Float64:
		var valFloat float64
		if len(values) > 0 {
			valFloat, err = strconv.ParseFloat(values[0], 64)
			if err != nil {
				return
			}
		}
		if valFloat == 0 {
			if name.required {
				return errors.New(h.handle.api.lang.Required(desc))
			}
			return
		}
		if err = h.validFloat64(valFloat, desc, field); err != nil {
			return
		}
		value.Set(reflect.ValueOf(valFloat).Convert(value.Type()))
	case reflect.Bool:
		var valBool bool
		if len(values) > 0 {
			valBool, err = strconv.ParseBool(values[0])
			if err != nil {
				return
			}
		}
		if field.tag.enum != nil && !inArrayAny(any(valBool), field.tag.enum) {
			return errors.New(h.handle.api.lang.Enum(desc, field.tag.enum))
		}
		value.Set(reflect.ValueOf(valBool).Convert(value.Type()))
	default:
	}
	return
}

func (h *handlerServer) handleParamByOther(ctx *Context, value reflect.Value) {
	for value.Kind() == reflect.Ptr {
		if value.Type().ConvertibleTo(typeContext) {
			value.Set(reflect.ValueOf(ctx).Convert(value.Type()))
			return
		}
		initPtr(value)
		value = value.Elem()
	}
}

func (h *handlerServer) validFloat64(vFloat float64, desc string, field *paramField) (err error) {
	if field.tag.lt != nil && vFloat >= *field.tag.lt {
		return errors.New(h.handle.api.lang.Lt(desc, *field.tag.lt))
	}
	if field.tag.lte != nil && vFloat > *field.tag.lte {
		return errors.New(h.handle.api.lang.Lte(desc, *field.tag.lte))
	}
	if field.tag.gt != nil && vFloat <= *field.tag.gt {
		return errors.New(h.handle.api.lang.Gt(desc, *field.tag.gt))
	}
	if field.tag.gte != nil && vFloat < *field.tag.gte {
		return errors.New(h.handle.api.lang.Gte(desc, *field.tag.gte))
	}
	if field.tag.multiple != nil {
		if *field.tag.multiple == 0 {
			return errors.New(h.handle.api.lang.MultipleOf(desc, *field.tag.multiple))
		}
		rs, _ := decimal.NewFromFloat(vFloat).Div(decimal.NewFromFloat(*field.tag.multiple)).Float64()
		if rs != float64(int64(rs)) {
			return errors.New(h.handle.api.lang.MultipleOf(desc, *field.tag.multiple))
		}
	}
	if field.tag.enum != nil && !inArrayAny(any(vFloat), field.tag.enum) {
		return errors.New(h.handle.api.lang.Enum(desc, field.tag.enum))
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
	value = h.removeMorPtrValue(value)
	newValue := value
	if newValue.Kind() != reflect.Ptr {
		newValue = reflect.New(value.Type())
	}
	if err = mediaType.Unmarshaler(reader, newValue); err != nil {
		return
	}
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	value.Set(newValue.Elem())
	return
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
				if val := tree.root.getValue(ctx.Request.URL.Path); val.handler != nil {
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
		code := http.StatusMovedPermanently
		if ctx.Request.Method != http.MethodGet {
			code = http.StatusTemporaryRedirect
		}
		http.Redirect(ctx.Writer, ctx.Request, tsrPath, code)
	})
	ctx.Next()
}

func (h *handlerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.pool.Get().(*Context)
	ctx.log = h.log
	ctx.writermem.reset(w)
	ctx.reset()
	ctx.Request = r
	if ctx.handleExcept == nil {
		ctx.handleExcept = h.handleExcept
	}
	h.generateRequestID(ctx)
	h.handleHTTPRequest(ctx)
	h.pool.Put(ctx)
}

func (h *handlerServer) generateRequestID(ctx *Context) {
	if h.handle.api.GenerateRequestID {
		pk, _ := uuid.NewV4()
		ctx.RequestID = pk.String()
	}
}

func (h *handlerServer) handleLogger(ctx *Context) {
	if ctx.log == nil {
		return
	}
	if _, ok := getFnByCovertInterface[LoggerContext](ctx.log); ok {
		newLog := h.copyLogger(ctx.log)
		if fn, fnOk := newLog.(LoggerContext); fnOk {
			fn.SetContext(ctx)
		}
		ctx.log = newLog
	}
}

func (h *handlerServer) copyLogger(log Logger) Logger {
	val := reflect.ValueOf(log)
	var newVal reflect.Value
	if val.Kind() == reflect.Ptr {
		newVal = reflect.New(val.Type().Elem())
	} else {
		newVal = reflect.New(val.Type()).Elem()
	}
	h.copyStruct(newVal, val)
	return newVal.Interface().(Logger)
}

func (h *handlerServer) copyStruct(dst, src reflect.Value) {
	if dst.Type() != src.Type() || src.IsZero() {
		return
	}
	switch src.Kind() {
	case reflect.Ptr:
		h.copyStruct(dst.Elem(), src.Elem())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		dst.Set(src)
	case reflect.Slice, reflect.Array:
		for i := 0; i < src.Len(); i++ {
			h.copyStruct(dst.Index(i), src.Index(i))
		}
	case reflect.Map:
		keys := src.MapKeys()
		for _, key := range keys {
			dst.SetMapIndex(key, src.MapIndex(key))
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			h.copyStruct(dst.Field(i), src.Field(i))
		}
	default:
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
	value := root.getValue(ctx.Request.URL.Path)
	if value.handler != nil {
		ctx.Params = value.params
		ctx.fullPath = value.fullPath
		value.handler(ctx)
		return
	}
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
	contentTypeList := strings.Split(ctx.Request.Header.Get("Content-Type"), ";")
	return MediaType(contentTypeList[0])
}

func (h *handlerServer) getResponseMediaType(ctx *Context) MediaType {
	if len(h.handle.api.responseMediaTypes) == 1 {
		return h.handle.api.responseMediaTypes[0]
	}
	mediaType := MediaType(ctx.Request.URL.Query().Get(returnMediaTypeField))
	if mediaType == "" || mediaType.Tag() == "" || !inArray(mediaType.MediaType(), h.handle.api.responseMediaTypes) {
		return h.handle.api.responseMediaTypes[0]
	}
	return mediaType.MediaType()
}
