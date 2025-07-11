package goapi

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/goodluckxu-go/goapi/response"
	json "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

func newHandlerServer(
	api *API,
	handle *handler,
) *handlerServer {
	return &handlerServer{
		api:    api,
		handle: handle,
	}
}

type handlerServer struct {
	api    *API
	handle *handler
}

func (h *handlerServer) Handle() {
	for _, static := range h.handle.statics {
		h.handleStatic(static)
	}
	for _, path := range h.handle.paths {
		for _, method := range path.methods {
			h.handlePaths(method, path, path.middlewares)
		}
	}
}

func (h *handlerServer) HttpHandler() http.Handler {
	for _, router := range h.api.routers {
		if err := h.api.mux.AddRouter(router.method, router.path, router.handler); err != nil {
			log.Fatal(err)
		}
	}
	for _, static := range h.api.statics {
		if err := h.api.mux.Static(static.path, static.handler); err != nil {
			log.Fatal(err)
		}
	}
	h.api.mux.NodFind(h.handleNodFind)
	h.api.mux.MethodNotAllowed(h.handleMethodNotAllowed)
	return h.api.mux
}

func (h *handlerServer) handleStatic(static *staticInfo) {
	root, _ := filepath.Abs(static.root)
	h.api.statics = append(h.api.statics, &appRouter{
		path:     static.path,
		isPrefix: true,
		method:   http.MethodGet,
		handler: func(ctx *Context) {
			ctx.middlewares = h.handle.defaultMiddlewares
			ctx.log = h.api.log
			ctx.middlewares = append(ctx.middlewares, func(ctx *Context) {
				name := strings.TrimPrefix(ctx.Request.URL.Path, static.path)
				http.ServeFile(ctx.Writer, ctx.Request, filepath.Join(root, name))
			})
			ctx.Next()
		},
		pos: root + fmt.Sprintf(" (fs) (%v Middleware)", len(h.handle.defaultMiddlewares)),
	})
}

func (h *handlerServer) handlePaths(method string, path *pathInfo, middlewares []Middleware) {
	h.api.routers = append(h.api.routers, &appRouter{
		path:   path.path,
		method: method,
		handler: func(ctx *Context) {
			h.handlePath(ctx, path)
		},
		pos: fmt.Sprintf("%v (%v Middleware)", path.pos, len(middlewares)),
	})
}

func (h *handlerServer) handleNodFind(ctx *Context) {
	mediaType := ctx.Request.URL.Query().Get("media_type")
	if (mediaType != jsonType && mediaType != xmlType) || len(h.api.responseMediaTypes) == 1 {
		mediaType = mediaTypeToTypeMap[h.api.responseMediaTypes[0]]
	}
	ctx.log = h.api.log
	ctx.mediaType = mediaType
	ctx.handleServer = h
	ctx.middlewares = append(h.handle.publicMiddlewares, func(ctx *Context) {
		http.NotFound(ctx.Writer, ctx.Request)
	})
	ctx.Next()
}

func (h *handlerServer) handleMethodNotAllowed(ctx *Context) {
	mediaType := ctx.Request.URL.Query().Get("media_type")
	if (mediaType != jsonType && mediaType != xmlType) || len(h.api.responseMediaTypes) == 1 {
		mediaType = mediaTypeToTypeMap[h.api.responseMediaTypes[0]]
	}
	ctx.log = h.api.log
	ctx.mediaType = mediaType
	ctx.handleServer = h
	ctx.middlewares = append(h.handle.publicMiddlewares, func(ctx *Context) {
		ctx.Writer.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = ctx.Writer.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
	})
	ctx.Next()
}

func (h *handlerServer) handlePath(ctx *Context, path *pathInfo) {
	mediaType := ctx.Request.URL.Query().Get("media_type")
	if (mediaType != jsonType && mediaType != xmlType) || len(h.api.responseMediaTypes) == 1 {
		mediaType = mediaTypeToTypeMap[h.api.responseMediaTypes[0]]
	}
	ctx.log = h.api.log
	ctx.mediaType = mediaType
	ctx.handleServer = h
	ctx.middlewares = append(path.middlewares, func(ctx *Context) {
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
		inputFields, err := h.handleInputFields(ctx, path.inTypes[len(path.inTypes)-1], path.inputFields)
		inputs[lastInputIdx] = inputFields
		if err != nil {
			response.HTTPException(validErrorCode, err.Error())
		}
		rs := path.funcValue.Call(inputs)
		if len(rs) != 1 {
			return
		}
		if resp, ok := rs[0].Interface().(Response); ok {
			resp.SetContentType(string(typeToMediaTypeMap[mediaType]))
			resp.Write(ctx.Writer)
			return
		}
		httpRes := &response.HTTPResponse[any]{
			HttpCode: 200,
			Header: map[string][]string{
				"Content-Type": {string(typeToMediaTypeMap[mediaType])},
			},
			Body: rs[0].Interface(),
		}
		httpRes.Write(ctx.Writer)
	})
	ctx.Next()
}

func (h *handlerServer) handleException(writer http.ResponseWriter, err any, mediaType string) {
	errStr := fmt.Sprintf("%v", err)
	var res exceptInfo
	err = json.Unmarshal([]byte(errStr), &res)
	var exceptRes Response
	if err != nil {
		exceptRes = h.api.exceptFunc(http.StatusInternalServerError, errStr)
		h.api.log.Error("panic: %v [recovered]\n%v", errStr, string(debug.Stack()))
	} else {
		exceptRes = h.api.exceptFunc(res.HttpCode, res.Detail)
	}
	exceptRes.SetContentType(string(typeToMediaTypeMap[mediaType]))
	exceptRes.Write(writer)
}

func (h *handlerServer) handleInputFields(ctx *Context, inputTypes reflect.Type, fields []fieldInfo) (inputValue reflect.Value, err error) {
	inputValue = reflect.New(inputTypes).Elem()
	var formType MediaType
	for _, field := range fields {
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
	var securityApiKeyList []reflect.Value
	securityApiKeyMap := map[string]struct{}{}
	for _, field := range fields {
		switch field.inType {
		case inTypeHeader:
			if err = h.handleHeader(ctx.Request, inputValue, field); err != nil {
				return
			}
		case inTypeCookie:
			if err = h.handleCookie(ctx.Request, inputValue, field); err != nil {
				return
			}
		case inTypeQuery:
			values := ctx.Request.URL.Query()[field.inTypeVal]
			if err = h.handleValue(inputValue, field, values); err != nil {
				return
			}
		case inTypePath:
			values := h.handleValueToValues(field._type, ctx.paths[field.inTypeVal])
			if err = h.handleValue(inputValue, field, values); err != nil {
				return
			}
		case inTypeForm:
			value := ""
			switch formType {
			case formUrlencoded:
				value = ctx.Request.Form.Get(field.inTypeVal)
			case formMultipart:
				if ctx.Request.MultipartForm != nil && ctx.Request.MultipartForm.Value[field.inTypeVal] != nil {
					value = ctx.Request.MultipartForm.Value[field.inTypeVal][0]
				}
			}
			values := h.handleValueToValues(field._type, value)
			if err = h.handleValue(inputValue, field, values); err != nil {
				return
			}
		case inTypeFile:
			var files []*multipart.FileHeader
			if ctx.Request.MultipartForm != nil {
				files = ctx.Request.MultipartForm.File[field.inTypeVal]
			}
			if files != nil {
				childField := h.getChildFieldVal(inputValue, field.deepIdx)
				switch childField.Type() {
				case typeFile:
					childField.Set(reflect.ValueOf(files[0]))
				case typeFiles:
					childField.Set(reflect.ValueOf(files))
				}
			}
		case inTypeBody:
			name := field.tag.desc
			if name == "" {
				name = field.name
			}
			if err = h.validLen(int(ctx.Request.ContentLength), name, field.tag); err != nil {
				return
			}
			if field._type.Implements(interfaceIoReadCloser) {
				if inArray("application/octet-stream", field.mediaTypes) {
					childField := h.getChildFieldVal(inputValue, field.deepIdx)
					childField.Set(reflect.ValueOf(ctx.Request.Body))
					continue
				}
				body := new(bytes.Buffer)
				buf := make([]byte, 512)
				nr, er := ctx.Request.Body.Read(buf)
				if nr > 0 {
					body.Write(buf[:nr])
					relContentType := strings.SplitN(http.DetectContentType(buf[:nr]), ";", 2)[0]
					if !inArray(MediaType(relContentType), field.mediaTypes) {
						err = fmt.Errorf(h.api.lang.ContentTypeNotSupported(relContentType))
						return
					}
				}
				readBody := func() io.ReadCloser {
					if er == nil {
						_, _ = io.Copy(body, ctx.Request.Body)
					}
					return io.NopCloser(body)
				}
				ctx.Request.Body = readBody()
				childField := h.getChildFieldVal(inputValue, field.deepIdx)
				childField.Set(reflect.ValueOf(ctx.Request.Body))
				continue
			}
			if bodyBytes, er := io.ReadAll(ctx.Request.Body); er == nil {
				if len(bodyBytes) == 0 {
					err = fmt.Errorf("%v", h.api.lang.Required(name))
					return
				}
				childField := h.getChildFieldVal(inputValue, field.deepIdx)
				if err = h.setBody(ctx.Request, childField, bodyBytes); err != nil {
					return
				}
			} else {
				err = fmt.Errorf("%v", h.api.lang.Required(name))
				return
			}
		case inTypeSecurityHTTPBearer:
			authorization := ctx.Request.Header.Get("Authorization")
			authList := strings.Split(authorization, " ")
			token := ""
			if len(authList) == 2 && authList[0] == "Bearer" {
				token = authList[1]
			}
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			h.handleSecurityDefaultParam(ctx, childField)
			security := childField.Interface().(HTTPBearer)
			security.HTTPBearer(token)
		case inTypeSecurityHTTPBearerJWT:
			authorization := ctx.Request.Header.Get("Authorization")
			authList := strings.Split(authorization, " ")
			token := ""
			if len(authList) == 2 && authList[0] == "Bearer" {
				token = authList[1]
			}
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			h.handleSecurityDefaultParam(ctx, childField)
			security := childField.Interface().(HTTPBearerJWT)
			jwt := &JWT{}
			if err = decryptJWT(jwt, token, security); err != nil {
				err = fmt.Errorf("%v", h.api.lang.JwtTranslate(err.Error()))
				return
			}
			security.HTTPBearerJWT(jwt)
		case inTypeSecurityHTTPBasic:
			username, password, _ := ctx.Request.BasicAuth()
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			h.handleSecurityDefaultParam(ctx, childField)
			security := childField.Interface().(HTTPBasic)
			security.HTTPBasic(username, password)
		case inTypeSecurityApiKey:
			pChildField := h.getChildFieldVal(inputValue, field.deepIdx[:len(field.deepIdx)-1])
			if _, ok := securityApiKeyMap[pChildField.String()]; !ok {
				securityApiKeyMap[pChildField.String()] = struct{}{}
				securityApiKeyList = append(securityApiKeyList, pChildField)
			}
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			switch field.inTypeSecurity {
			case inTypeHeader:
				if err = h.handleHeader(ctx.Request, inputValue, field); err != nil {
					return
				}
			case inTypeCookie:
				if err = h.handleCookie(ctx.Request, inputValue, field); err != nil {
					return
				}
			case inTypeQuery:
				values := ctx.Request.URL.Query()[field.inTypeVal]
				if err = h.handleValue(inputValue, field, values); err != nil {
					return
				}
			}
		}
	}
	for _, securityApiKey := range securityApiKeyList {
		if securityApiKey.IsValid() {
			h.handleSecurityDefaultParam(ctx, securityApiKey)
			security := securityApiKey.Interface().(ApiKey)
			security.ApiKey()
		}
	}
	return
}

func (h *handlerServer) handleSecurityDefaultParam(ctx *Context, securityValue reflect.Value) {
	numField := securityValue.Elem().NumField()
	for i := 0; i < numField; i++ {
		name := securityValue.Elem().Type().Field(i).Name
		if name[0] < 'A' || name[0] > 'Z' {
			continue
		}
		switch securityValue.Elem().Field(i).Type() {
		case typeContext:
			securityValue.Elem().Field(i).Set(reflect.ValueOf(ctx))
		}
	}
}

func (h *handlerServer) handleHeader(req *http.Request, inputValue reflect.Value, field fieldInfo) (err error) {
	values := h.handleValueToValues(field._type, req.Header.Get(field.inTypeVal))
	if err = h.handleValue(inputValue, field, values); err != nil {
		return
	}
	return
}

func (h *handlerServer) handleCookie(req *http.Request, inputValue reflect.Value, field fieldInfo) (err error) {
	name := strings.TrimSuffix(field.tag.desc, "Read the value of document.cookie")
	if name == "" {
		name = field.inTypeVal
	}
	cookie, er := req.Cookie(field.inTypeVal)
	if er != nil || cookie.Value == "" {
		if field.required {
			err = fmt.Errorf("%v", h.api.lang.Required(name))
			return
		}
		return
	}
	if field._type == typeCookie {
		if err = h.validString(cookie.Value, name, field.tag); err != nil {
			return
		}
		childVal := h.getChildFieldVal(inputValue, field.deepIdx)
		childVal.Set(reflect.ValueOf(cookie))
	} else {
		values := h.handleValueToValues(field._type, cookie.Value)
		if err = h.handleValue(inputValue, field, values); err != nil {
			return
		}
	}
	return
}

func (h *handlerServer) handleValueToValues(fType reflect.Type, value string) (values []string) {
	if value != "" {
		for fType.Kind() == reflect.Ptr {
			fType = fType.Elem()
		}
		values = []string{value}
		if fType.Kind() == reflect.Slice {
			values = []string{}
			valList := strings.Split(value, ",")
			for _, v := range valList {
				values = append(values, strings.TrimSpace(v))
			}
		}
	}
	return
}

func (h *handlerServer) handleValue(inputValue reflect.Value, field fieldInfo, values []string) (err error) {
	name := field.tag.desc
	if name == "" {
		name = field.inTypeVal
	}
	required := field.required
	if values == nil {
		if required {
			err = fmt.Errorf("%v", h.api.lang.Required(name))
		}
		return
	}
	fType := field._type
	for fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}
	if fType.Kind() == reflect.Slice {
		if len(values) == 0 {
			if required {
				err = fmt.Errorf("%v", h.api.lang.Required(name))
			}
			return
		}
	} else {
		if len(values) == 0 || values[0] == "" {
			if required {
				err = fmt.Errorf("%v", h.api.lang.Required(name))
			}
			return
		}
	}
	childVal := h.getChildFieldVal(inputValue, field.deepIdx)
	return h.setValue(childVal, values, name, field.tag)
}

func (h *handlerServer) setValue(fVal reflect.Value, values []string, name string, tag *fieldTagInfo) (err error) {
	if len(values) == 0 {
		return
	}
	switch fVal.Kind() {
	case reflect.String:
		if err = h.validString(values[0], name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(values[0]))
	case reflect.Int:
		var v int64
		if v, err = strconv.ParseInt(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(int(v)))
	case reflect.Int8:
		var v int64
		if v, err = strconv.ParseInt(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(int8(v)))
	case reflect.Int16:
		var v int64
		if v, err = strconv.ParseInt(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(int16(v)))
	case reflect.Int32:
		var v int64
		if v, err = strconv.ParseInt(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(int32(v)))
	case reflect.Int64:
		var v int64
		if v, err = strconv.ParseInt(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(v))
	case reflect.Uint:
		var v uint64
		if v, err = strconv.ParseUint(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(uint(v)))
	case reflect.Uint8:
		var v uint64
		if v, err = strconv.ParseUint(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(uint8(v)))
	case reflect.Uint16:
		var v uint64
		if v, err = strconv.ParseUint(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(uint16(v)))
	case reflect.Uint32:
		var v uint64
		if v, err = strconv.ParseUint(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(uint32(v)))
	case reflect.Uint64:
		var v uint64
		if v, err = strconv.ParseUint(values[0], 10, 64); err != nil {
			return
		}
		if err = h.validFloat64(float64(v), name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(v))
	case reflect.Float32:
		var v float64
		if v, err = strconv.ParseFloat(values[0], 64); err != nil {
			return
		}
		if err = h.validFloat64(v, name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(float32(v)))
	case reflect.Float64:
		var v float64
		if v, err = strconv.ParseFloat(values[0], 64); err != nil {
			return
		}
		if err = h.validFloat64(v, name, tag); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(v))
	case reflect.Bool:
		var v bool
		if v, err = strconv.ParseBool(values[0]); err != nil {
			return
		}
		fVal.Set(reflect.ValueOf(v))
	case reflect.Ptr:
		h.initPtr(fVal)
		if err = h.setValue(fVal.Elem(), values, name, tag); err != nil {
			return
		}
	case reflect.Slice:
		if err = h.validLen(len(values), name, tag); err != nil {
			return
		}
		if tag.unique {
			valCount := map[string]int{}
			for _, val := range values {
				valCount[val]++
			}
			for _, count := range valCount {
				if count > 1 {
					err = fmt.Errorf("%v", h.api.lang.Unique(name))
					return
				}
			}
		}
		list := reflect.MakeSlice(fVal.Type(), len(values), len(values))
		for key, val := range values {
			if err = h.setValue(list.Index(key), []string{val}, name, &fieldTagInfo{}); err != nil {
				return
			}
		}
		fVal.Set(list)
	default:
	}
	return
}

func (h *handlerServer) validFloat64(f float64, name string, tag *fieldTagInfo) (err error) {
	if tag.lt != nil && f >= *tag.lt {
		err = fmt.Errorf("%v", h.api.lang.Lt(name, *tag.lt))
		return
	}
	if tag.lte != nil && f > *tag.lte {
		err = fmt.Errorf("%v", h.api.lang.Lte(name, *tag.lte))
		return
	}
	if tag.gt != nil && f <= *tag.gt {
		err = fmt.Errorf("%v", h.api.lang.Gt(name, *tag.gt))
		return
	}
	if tag.gte != nil && f < *tag.gte {
		err = fmt.Errorf("%v", h.api.lang.Gte(name, *tag.gte))
		return
	}
	if tag.multiple != nil {
		if *tag.multiple == 0 {
			err = fmt.Errorf("%v", h.api.lang.MultipleOf(name, *tag.multiple))
			return
		}
		rs, _ := decimal.NewFromFloat(f).Div(decimal.NewFromFloat(*tag.multiple)).Float64()
		if rs != float64(int64(rs)) {
			err = fmt.Errorf("%v", h.api.lang.MultipleOf(name, *tag.multiple))
			return
		}
	}
	var enum []float64
	for _, v := range tag.enum {
		enum = append(enum, v.(float64))
	}
	if len(enum) > 0 && !inArray(f, enum) {
		err = fmt.Errorf("%v", h.api.lang.Enum(name, tag.enum))
		return
	}
	return
}

func (h *handlerServer) validString(s string, name string, tag *fieldTagInfo) (err error) {
	if err = h.validLen(len(s), name, tag); err != nil {
		return
	}
	if tag.regexp != "" && !regexp.MustCompile(tag.regexp).MatchString(s) {
		err = fmt.Errorf("%v", h.api.lang.Regexp(name, tag.regexp))
		return
	}
	var enum []string
	for _, v := range tag.enum {
		enum = append(enum, v.(string))
	}
	if len(enum) > 0 && !inArray(s, enum) {
		err = fmt.Errorf("%v", h.api.lang.Enum(name, tag.enum))
		return
	}
	return
}

func (h *handlerServer) validLen(l int, name string, tag *fieldTagInfo) (err error) {
	if tag.min > 0 && uint64(l) < tag.min {
		err = fmt.Errorf("%v", h.api.lang.Min(name, tag.min))
		return
	}
	if tag.max != nil && uint64(l) > *tag.max {
		err = fmt.Errorf("%v", h.api.lang.Max(name, *tag.max))
		return
	}
	return
}

func (h *handlerServer) setBody(req *http.Request, fVal reflect.Value, body []byte) (err error) {
	h.initPtr(fVal)
	newVal := fVal
	if newVal.Kind() != reflect.Ptr {
		newVal = reflect.New(newVal.Type())
	}
	mediaType := mediaTypeToTypeMap[MediaType(req.Header.Get("Content-Type"))]
	switch mediaType {
	case jsonType:
		if err = json.Unmarshal(body, newVal.Interface()); err != nil {
			return
		}
	case xmlType:
		if err = xml.Unmarshal(body, newVal.Interface()); err != nil {
			return
		}
	default:
		if typeBytes == fVal.Type() {
			fVal.Set(reflect.ValueOf(body))
			return
		}
		switch fVal.Kind() {
		case reflect.String:
			fVal.Set(reflect.ValueOf(string(body)))
		default:
		}
		return
	}
	if err = h.validBody(newVal, mediaType); err != nil {
		return
	}
	if fVal.Kind() != reflect.Ptr {
		newVal = newVal.Elem()
	}
	fVal.Set(newVal)
	return
}

func (h *handlerServer) validBody(val reflect.Value, mediaType string) (err error) {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Map:
		keys := val.MapKeys()
		for _, key := range keys {
			if err = h.validBody(val.MapIndex(key), mediaType); err != nil {
				return
			}
		}
	case reflect.Slice:
		vLen := val.Len()
		for i := 0; i < vLen; i++ {
			if err = h.validBody(val.Index(i), mediaType); err != nil {
				return
			}
		}
	case reflect.Struct:
		key := fmt.Sprintf("%v.%v", val.Type().PkgPath(), val.Type().Name())
		sInfo := h.handle.structs[key]
		for _, field := range sInfo.fields {
			name := field.tag.desc
			if name == "" {
				name = field.name
			}
			myFName := field.fieldMap[typeToMediaTypeMap[mediaType]]
			v := val.Field(field.deepIdx[0])
			fType := field._type
			for fType.Kind() == reflect.Ptr {
				fType = fType.Elem()
				v = v.Elem()
			}
			switch field._type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				vFloat := float64(v.Int())
				if vFloat == 0 {
					if myFName.required {
						err = fmt.Errorf("%v", h.api.lang.Required(name))
						return
					}
					continue
				}
				if err = h.validFloat64(vFloat, name, field.tag); err != nil {
					return
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				vFloat := float64(v.Uint())
				if vFloat == 0 {
					if myFName.required {
						err = fmt.Errorf("%v", h.api.lang.Required(name))
						return
					}
					continue
				}
				if err = h.validFloat64(vFloat, name, field.tag); err != nil {
					return
				}
			case reflect.Float32, reflect.Float64:
				vFloat := v.Float()
				if vFloat == 0 {
					if myFName.required {
						err = fmt.Errorf("%v", h.api.lang.Required(name))
						return
					}
					continue
				}
				if err = h.validFloat64(vFloat, name, field.tag); err != nil {
					return
				}
			case reflect.String:
				vStr := v.String()
				if v.String() == "" {
					if myFName.required {
						err = fmt.Errorf("%v", h.api.lang.Required(name))
						return
					}
					continue
				}
				if err = h.validString(vStr, name, field.tag); err != nil {
					return
				}
			case reflect.Slice:
				vLen := v.Len()
				if vLen == 0 {
					if myFName.required {
						err = fmt.Errorf("%v", h.api.lang.Required(name))
						return
					}
					continue
				}
				if err = h.validLen(vLen, name, field.tag); err != nil {
					return
				}
				if field.tag.unique {
					valCount := map[reflect.Value]int{}
					for i := 0; i < vLen; i++ {
						valCount[v.Index(i)]++
					}
					for _, count := range valCount {
						if count > 1 {
							err = fmt.Errorf("%v", h.api.lang.Unique(name))
							return
						}
					}
				}
				if err = h.validBody(v, mediaType); err != nil {
					return
				}
			case reflect.Map:
				vLen := len(v.MapKeys())
				if vLen == 0 {
					if myFName.required {
						err = fmt.Errorf("%v", h.api.lang.Required(name))
						return
					}
					continue
				}
				if err = h.validLen(vLen, name, field.tag); err != nil {
					return
				}
				if err = h.validBody(v, mediaType); err != nil {
					return
				}
			case reflect.Struct:
				if err = h.validBody(v, mediaType); err != nil {
					return
				}
			default:
			}
		}
	default:
	}
	return
}

func (h *handlerServer) getChildFieldVal(inputFiled reflect.Value, deepIdx []int) (childField reflect.Value) {
	if len(deepIdx) == 0 {
		return
	}
	childField = inputFiled.Field(deepIdx[0])
	if len(deepIdx) > 1 {
		for childField.Kind() == reflect.Ptr {
			if isFixedType(childField.Type()) {
				break
			}
			h.initPtr(childField)
			childField = childField.Elem()
		}
	}
	for _, idx := range deepIdx[1:] {
		for childField.Kind() == reflect.Ptr {
			if isFixedType(childField.Type()) {
				break
			}
			h.initPtr(childField)
			childField = childField.Elem()
		}
		childField = childField.Field(idx)
	}
	return
}

func (h *handlerServer) initPtr(fVal reflect.Value) {
	if fVal.Kind() != reflect.Ptr || !fVal.IsNil() {
		return
	}
	newVal := reflect.New(fVal.Type().Elem())
	fVal.Set(newVal)
}
