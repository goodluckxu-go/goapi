package goapi

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
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
	api     *API
	handle  *handler
	pathMap map[string]string
}

func (h *handlerServer) Handle() {
	for _, static := range h.handle.statics {
		h.handleStatic(static)
	}
	for _, path := range h.handle.paths {
		for _, method := range path.methods {
			h.handlePaths(method, path)
		}
	}
	h.api.app.Handle(func(ctx *Context) {
		ctx = &Context{
			Request: ctx.Request,
			Writer:  &ResponseWriter{ResponseWriter: ctx.Writer},
		}
		isFind := false
		for _, router := range h.api.routers {
			if router.method == ctx.Request.Method {
				if router.isPrefix && strings.HasPrefix(ctx.Request.URL.Path, router.path) {
					router.handler(ctx)
					isFind = true
				} else if _, err := h.getPaths(router.path, ctx.Request.URL.Path); err == nil {
					router.handler(ctx)
					isFind = true
				}
				if isFind {
					break
				}
			}
		}
		if !isFind {
			done := make(chan struct{})
			go h.handlePath(ctx, nil, done)
			<-done
		}
	})
}

func (h *handlerServer) handleStatic(static staticInfo) {
	root, _ := filepath.Abs(static.root)
	h.api.routers = append(h.api.routers, appRouter{
		path:     static.path,
		isPrefix: true,
		method:   http.MethodGet,
		handler: func(ctx *Context) {
			ctx.middlewares = h.handle.middlewares
			ctx.log = h.api.log
			ctx.routerFunc = func(done chan struct{}) {
				name := strings.TrimPrefix(ctx.Request.URL.Path, static.path)
				http.ServeFile(ctx.Writer, ctx.Request, filepath.Join(root, name))
				done <- struct{}{}
			}
			ctx.Next()
		},
	})
}

func (h *handlerServer) handlePaths(method string, path pathInfo) {
	h.api.routers = append(h.api.routers, appRouter{
		path:   path.path,
		method: method,
		handler: func(ctx *Context) {
			done := make(chan struct{})
			go h.handlePath(ctx, &path, done)
			<-done
		},
	})
}

func (h *handlerServer) handlePath(ctx *Context, path *pathInfo, done chan struct{}) {
	ctx.log = h.api.log
	mediaType := ctx.Request.URL.Query().Get("media_type")
	if (mediaType != jsonType && mediaType != xmlType) || len(h.api.responseMediaTypes) == 1 {
		mediaType = mediaTypeToTypeMap[h.api.responseMediaTypes[0]]
	}
	defer func() {
		if er := recover(); er != nil {
			h.handleException(ctx.Writer, er, mediaType)
			done <- struct{}{}
		}
	}()
	httpRes := &HTTPResponse[any]{
		HttpCode: 200,
		Header: map[string]string{
			"Content-Type": string(typeToMediaTypeMap[mediaType]),
		},
	}
	if path == nil {
		ctx.middlewares = append([]Middleware{notFind()}, h.handle.middlewares...)
		ctx.Next()
		done <- struct{}{}
		return
	}
	h.pathMap, _ = h.getPaths(path.path, ctx.Request.URL.Path)
	var inputs []reflect.Value
	ctx.middlewares = path.middlewares
	if len(path.inTypes) == 2 {
		inputs = append(inputs, reflect.ValueOf(ctx))
	}
	inputFields, err := h.handleInputFields(ctx.Request, path.inTypes[len(path.inTypes)-1], path.inputFields)
	inputs = append(inputs, inputFields)
	ctx.routerFunc = func(done chan struct{}) {
		defer func() {
			if er := recover(); er != nil {
				h.handleException(ctx.Writer, er, mediaType)
				done <- struct{}{}
			}
		}()
		if err != nil {
			HTTPException(validErrorCode, err.Error())
		}
		rs := path.funcValue.Call(inputs)
		if len(rs) != 1 {
			done <- struct{}{}
			return
		}
		if rs[0].Type().Implements(typeResponse) {
			resp := rs[0].Interface().(Response)
			httpRes.HttpCode = resp.GetHttpCode()
			for k, v := range resp.GetHeaders() {
				httpRes.Header[k] = v
			}
			httpRes.Body = resp.GetBody()
		} else {
			httpRes.Body = rs[0].Interface()
		}
		for k, v := range httpRes.Header {
			ctx.Writer.Header().Set(k, v)
		}
		ctx.Writer.WriteHeader(httpRes.GetHttpCode())
		_, _ = ctx.Writer.Write(httpRes.Bytes())
		done <- struct{}{}
	}
	ctx.Next()
	done <- struct{}{}
}

func (h *handlerServer) handleException(writer http.ResponseWriter, err any, mediaType string) {
	httpRes := &HTTPResponse[any]{
		HttpCode: 200,
		Header: map[string]string{
			"Content-Type": string(typeToMediaTypeMap[mediaType]),
		},
	}
	errStr := fmt.Sprintf("%v", err)
	var res exceptInfo
	err = json.Unmarshal([]byte(errStr), &res)
	if err != nil {
		exceptRes := h.api.exceptFunc(http.StatusInternalServerError, errStr)
		httpRes.HttpCode = http.StatusInternalServerError
		httpRes.Body = exceptRes.GetBody()
		h.api.log.Error("panic: %v [recovered]\n%v", errStr, string(debug.Stack()))
	} else {
		exceptRes := h.api.exceptFunc(res.HttpCode, res.Detail)
		for k, v := range res.Header {
			httpRes.Header[k] = v
		}
		httpRes.HttpCode = exceptRes.GetHttpCode()
		httpRes.Body = exceptRes.GetBody()
	}
	for k, v := range httpRes.GetHeaders() {
		writer.Header().Set(k, v)
	}
	writer.WriteHeader(httpRes.GetHttpCode())
	_, _ = writer.Write(httpRes.Bytes())
}

func (h *handlerServer) handleInputFields(req *http.Request, inputTypes reflect.Type, fields []fieldInfo) (inputValue reflect.Value, err error) {
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
		if err = req.ParseForm(); err != nil {
			return
		}
	case formMultipart:
		if err = req.ParseMultipartForm(32 << 20); err != nil {
			return
		}
	}
	var securityApiKey reflect.Value
	for _, field := range fields {
		switch field.inType {
		case inTypeHeader:
			if err = h.handleHeader(req, inputValue, field); err != nil {
				return
			}
		case inTypeCookie:
			if err = h.handleCookie(req, inputValue, field); err != nil {
				return
			}
		case inTypeQuery:
			values := req.URL.Query()[field.inTypeVal]
			if err = h.handleValue(inputValue, field, values); err != nil {
				return
			}
		case inTypePath:
			values := h.handleValueToValues(field._type, h.pathMap[field.inTypeVal])
			if err = h.handleValue(inputValue, field, values); err != nil {
				return
			}
		case inTypeForm:
			value := ""
			switch formType {
			case formUrlencoded:
				value = req.Form.Get(field.inTypeVal)
			case formMultipart:
				if req.MultipartForm != nil && req.MultipartForm.Value[field.inTypeVal] != nil {
					value = req.MultipartForm.Value[field.inTypeVal][0]
				}
			}
			values := h.handleValueToValues(field._type, value)
			if err = h.handleValue(inputValue, field, values); err != nil {
				return
			}
		case inTypeFile:
			var files []*multipart.FileHeader
			if req.MultipartForm != nil {
				files = req.MultipartForm.File[field.inTypeVal]
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
			if bodyBytes, er := io.ReadAll(req.Body); er == nil {
				if len(bodyBytes) == 0 {
					err = fmt.Errorf(h.api.lang.Required("body"))
					return
				}
				childField := h.getChildFieldVal(inputValue, field.deepIdx)
				if err = h.setBody(req, childField, bodyBytes); err != nil {
					return
				}
			} else {
				err = fmt.Errorf(h.api.lang.Required("body"))
				return
			}
		case inTypeSecurityHTTPBearer:
			authorization := req.Header.Get("Authorization")
			authList := strings.Split(authorization, " ")
			token := ""
			if len(authList) == 2 && authList[0] == "Bearer" {
				token = authList[1]
			}
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			security := childField.Interface().(HTTPBearer)
			security.HTTPBearer(token)
		case inTypeSecurityHTTPBasic:
			username, password, _ := req.BasicAuth()
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			security := childField.Interface().(HTTPBasic)
			security.HTTPBasic(username, password)
		case inTypeSecurityApiKey:
			if !securityApiKey.IsValid() {
				securityApiKey = h.getChildFieldVal(inputValue, field.deepIdx[:len(field.deepIdx)-1])
			}
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			switch field.inTypeSecurity {
			case inTypeHeader:
				if err = h.handleHeader(req, inputValue, field); err != nil {
					return
				}
			case inTypeCookie:
				if err = h.handleCookie(req, inputValue, field); err != nil {
					return
				}
			case inTypeQuery:
				values := req.URL.Query()[field.inTypeVal]
				if err = h.handleValue(inputValue, field, values); err != nil {
					return
				}
			}
		}
	}
	if securityApiKey.IsValid() {
		security := securityApiKey.Interface().(ApiKey)
		security.ApiKey()
	}
	return
}

func (h *handlerServer) getPaths(path, urlPath string) (rs map[string]string, err error) {
	rs = map[string]string{}
	pathList := strings.Split(path, "/")
	relPathList := strings.Split(urlPath, "/")
	if len(pathList) != len(relPathList) {
		err = fmt.Errorf("path format error")
		rs = nil
		return
	}
	for k, v := range pathList {
		relV := relPathList[k]
		left := strings.Index(v, "{")
		right := strings.Index(v, "}")
		if left != -1 && right != -1 {
			right = len(v) - (right + 1)
			if v[:left] != relPathList[k][:left] || v[len(v)-right:] != relPathList[k][len(relV)-right:] {
				err = fmt.Errorf("path format error")
				rs = nil
				return
			}
			rs[v[left+1:len(v)-right-1]] = relPathList[k][left : len(relV)-right]
		} else if relV != v {
			err = fmt.Errorf("path format error")
			rs = nil
			return
		}
	}
	return
}

func (h *handlerServer) handleHeader(req *http.Request, inputValue reflect.Value, field fieldInfo) (err error) {
	values := h.handleValueToValues(field._type, req.Header.Get(field.inTypeVal))
	if err = h.handleValue(inputValue, field, values); err != nil {
		return
	}
	return
}

func (h *handlerServer) handleCookie(req *http.Request, inputValue reflect.Value, field fieldInfo) (err error) {
	cookie, er := req.Cookie(field.inTypeVal)
	if er != nil {
		return
	}
	if field._type == typeCookie {
		name := field.tag.desc
		if name == "" {
			name = field.inTypeVal
		}
		if er != nil || cookie.Value == "" {
			if field.required {
				err = fmt.Errorf(h.api.lang.Required(name))
				return
			}
			return
		}
		if err = h.validString(cookie.Value, name, field.tag); err != nil {
			return
		}
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
			err = fmt.Errorf(h.api.lang.Required(name))
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
				err = fmt.Errorf(h.api.lang.Required(name))
			}
			return
		}
	} else {
		if len(values) == 0 || values[0] == "" {
			if required {
				err = fmt.Errorf(h.api.lang.Required(name))
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
					err = fmt.Errorf(h.api.lang.Unique(name))
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
		err = fmt.Errorf(h.api.lang.Lt(name, *tag.lt))
		return
	}
	if tag.lte != nil && f > *tag.lte {
		err = fmt.Errorf(h.api.lang.Lte(name, *tag.lte))
		return
	}
	if tag.gt != nil && f <= *tag.gt {
		err = fmt.Errorf(h.api.lang.Gt(name, *tag.gt))
		return
	}
	if tag.gte != nil && f < *tag.gte {
		err = fmt.Errorf(h.api.lang.Gte(name, *tag.gte))
		return
	}
	if tag.multiple != nil {
		if *tag.multiple == 0 {
			err = fmt.Errorf(h.api.lang.MultipleOf(name, *tag.multiple))
			return
		}
		rs, _ := decimal.NewFromFloat(f).Div(decimal.NewFromFloat(*tag.multiple)).Float64()
		if rs != float64(int64(rs)) {
			err = fmt.Errorf(h.api.lang.MultipleOf(name, *tag.multiple))
			return
		}
	}
	var enum []float64
	for _, v := range tag.enum {
		enum = append(enum, v.(float64))
	}
	if len(enum) > 0 && !inArray(f, enum) {
		err = fmt.Errorf(h.api.lang.Enum(name, tag.enum))
		return
	}
	return
}

func (h *handlerServer) validString(s string, name string, tag *fieldTagInfo) (err error) {
	if err = h.validLen(len(s), name, tag); err != nil {
		return
	}
	if tag.regexp != "" && !regexp.MustCompile(tag.regexp).MatchString(s) {
		err = fmt.Errorf(h.api.lang.Regexp(name, tag.regexp))
		return
	}
	var enum []string
	for _, v := range tag.enum {
		enum = append(enum, v.(string))
	}
	if len(enum) > 0 && !inArray(s, enum) {
		err = fmt.Errorf(h.api.lang.Enum(name, tag.enum))
		return
	}
	return
}

func (h *handlerServer) validLen(l int, name string, tag *fieldTagInfo) (err error) {
	if tag.min > 0 && uint64(l) < tag.min {
		err = fmt.Errorf(h.api.lang.Min(name, tag.min))
		return
	}
	if tag.max != nil && uint64(l) > *tag.max {
		err = fmt.Errorf(h.api.lang.Max(name, *tag.max))
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
						err = fmt.Errorf(h.api.lang.Required(name))
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
						err = fmt.Errorf(h.api.lang.Required(name))
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
						err = fmt.Errorf(h.api.lang.Required(name))
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
						err = fmt.Errorf(h.api.lang.Required(name))
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
						err = fmt.Errorf(h.api.lang.Required(name))
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
							err = fmt.Errorf(h.api.lang.Unique(name))
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
						err = fmt.Errorf(h.api.lang.Required(name))
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
