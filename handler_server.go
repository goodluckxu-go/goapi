package goapi

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func newHandlerServer(
	lang Lang,
	app APP, paths []pathInfo,
	exceptFunc func(httpCode int, detail string) Response,
	structs map[string]*structInfo,
) *handlerServer {
	return &handlerServer{
		lang:       lang,
		app:        app,
		paths:      paths,
		exceptFunc: exceptFunc,
		structs:    structs,
	}
}

type handlerServer struct {
	lang       Lang
	app        APP
	paths      []pathInfo
	pathMap    map[string]string
	exceptFunc func(httpCode int, detail string) Response
	structs    map[string]*structInfo
}

func (h *handlerServer) Handle() {
	for _, path := range h.paths {
		for _, method := range path.methods {
			h.handlePaths(method, path)
		}
	}
}

func (h *handlerServer) handlePaths(method string, path pathInfo) {
	switch method {
	case http.MethodGet:
		h.app.GET(path.path, func(req *http.Request, writer http.ResponseWriter) {
			done := make(chan struct{})
			go h.handlePath(req, writer, path, done)
			<-done
		})
	case http.MethodPut:
		h.app.PUT(path.path, func(req *http.Request, writer http.ResponseWriter) {
			done := make(chan struct{})
			go h.handlePath(req, writer, path, done)
			<-done
		})
	case http.MethodPost:
		h.app.POST(path.path, func(req *http.Request, writer http.ResponseWriter) {
			done := make(chan struct{})
			go h.handlePath(req, writer, path, done)
			<-done
		})
	case http.MethodDelete:
		h.app.DELETE(path.path, func(req *http.Request, writer http.ResponseWriter) {
			done := make(chan struct{})
			go h.handlePath(req, writer, path, done)
			<-done
		})
	case http.MethodOptions:
		h.app.OPTIONS(path.path, func(req *http.Request, writer http.ResponseWriter) {
			done := make(chan struct{})
			go h.handlePath(req, writer, path, done)
			<-done
		})
	case http.MethodHead:
		h.app.HEAD(path.path, func(req *http.Request, writer http.ResponseWriter) {
			done := make(chan struct{})
			go h.handlePath(req, writer, path, done)
			<-done
		})
	case http.MethodPatch:
		h.app.PATCH(path.path, func(req *http.Request, writer http.ResponseWriter) {
			done := make(chan struct{})
			go h.handlePath(req, writer, path, done)
			<-done
		})
	case http.MethodTrace:
		h.app.TRACE(path.path, func(req *http.Request, writer http.ResponseWriter) {
			done := make(chan struct{})
			go h.handlePath(req, writer, path, done)
			<-done
		})
	}
}

func (h *handlerServer) handlePath(req *http.Request, writer http.ResponseWriter, path pathInfo, done chan struct{}) {
	mediaType := req.URL.Query().Get("media_type")
	if (mediaType != jsonType && mediaType != xmlType) || len(path.res.mediaTypes) == 1 {
		mediaType = path.res.mediaTypes[0]._type
	}
	httpRes := &HTTPResponse[any]{
		HttpCode: 200,
		Header: map[string]string{
			"Content-Type": string(typeToMediaTypeMap[mediaType]),
		},
	}
	defer func() {
		if err := recover(); err != nil {
			errStr := fmt.Sprintf("%v", err)
			var res exceptInfo
			err = json.Unmarshal([]byte(errStr), &res)
			if err != nil {
				exceptRes := h.exceptFunc(http.StatusInternalServerError, errStr)
				httpRes.HttpCode = http.StatusInternalServerError
				httpRes.Body = exceptRes.GetBody()
			} else {
				exceptRes := h.exceptFunc(res.HttpCode, res.Detail)
				for k, v := range res.Header {
					httpRes.Header[k] = v
				}
				httpRes.HttpCode = exceptRes.GetHttpCode()
				httpRes.Body = exceptRes.GetBody()
			}
			writer.WriteHeader(httpRes.GetHttpCode())
			for k, v := range httpRes.GetHeaders() {
				writer.Header().Set(k, v)
			}
			_, _ = writer.Write(httpRes.Bytes())
			done <- struct{}{}
		}
	}()
	h.pathMap, _ = h.getPaths(path.path, req.URL.Path)
	var inputs []reflect.Value
	ctx := &Context{
		Request:     req,
		Writer:      writer,
		middlewares: path.middlewares,
	}
	if len(path.inTypes) == 2 {
		inputs = append(inputs, reflect.ValueOf(ctx))
	}
	inputs = append(inputs, h.handleInputFields(req, path.inTypes[len(path.inTypes)-1], path.inputFields))
	ctx.routerFunc = func() {
		rs := path.funcValue.Call(inputs)
		for k, v := range httpRes.Header {
			writer.Header().Set(k, v)
		}
		if len(rs) == 1 {
			httpRes.Body = rs[0].Interface()
			_, _ = writer.Write(httpRes.Bytes())
		}
	}
	ctx.Next()
	done <- struct{}{}
	return
}

func (h *handlerServer) handleInputFields(req *http.Request, inputTypes reflect.Type, fields []fieldInfo) (inputValue reflect.Value) {
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
		if err := req.ParseForm(); err != nil {
			HTTPException(422, err.Error())
		}
	case formMultipart:
		if err := req.ParseMultipartForm(32 << 20); err != nil {
			HTTPException(422, err.Error())
		}
	}
	var securityApiKey reflect.Value
	for _, field := range fields {
		switch field.inType {
		case inTypeHeader:
			value := req.Header.Get(field.inTypeVal)
			var values []string
			if value != "" {
				fType := field._type
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
			if err := h.handleValue(inputValue, field, values); err != nil {
				HTTPException(validErrorCode, err.Error())
			}
		case inTypeCookie:
			cookie, err := req.Cookie(field.inTypeVal)
			if field._type == typeCookie {
				name := field.tag.desc
				if name == "" {
					name = field.inTypeVal
				}
				if err != nil || cookie.Value == "" {
					if field.mediaTypes[0].required {
						HTTPException(validErrorCode, h.lang.Required(name))
					}
					return
				}
				if err = h.validString(cookie.Value, name, field.tag); err != nil {
					HTTPException(validErrorCode, err.Error())
				}
			} else {
				var values []string
				if cookie.Value != "" {
					fType := field._type
					for fType.Kind() == reflect.Ptr {
						fType = fType.Elem()
					}
					values = []string{cookie.Value}
					if fType.Kind() == reflect.Slice {
						values = []string{}
						valList := strings.Split(cookie.Value, ",")
						for _, v := range valList {
							values = append(values, strings.TrimSpace(v))
						}
					}
				}
				if err = h.handleValue(inputValue, field, values); err != nil {
					HTTPException(validErrorCode, err.Error())
				}
			}
		case inTypeQuery:
			values := req.URL.Query()[field.inTypeVal]
			if err := h.handleValue(inputValue, field, values); err != nil {
				HTTPException(validErrorCode, err.Error())
			}
		case inTypePath:
			value := h.pathMap[field.inTypeVal]
			var values []string
			if value != "" {
				fType := field._type
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
			if err := h.handleValue(inputValue, field, values); err != nil {
				HTTPException(validErrorCode, err.Error())
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
			var values []string
			if value != "" {
				fType := field._type
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
			if err := h.handleValue(inputValue, field, values); err != nil {
				HTTPException(validErrorCode, err.Error())
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
			if bodyBytes, err := io.ReadAll(req.Body); err == nil {
				if len(bodyBytes) == 0 {
					HTTPException(validErrorCode, h.lang.Required("body"))
				}
				childField := h.getChildFieldVal(inputValue, field.deepIdx)
				if err = h.setBody(req, childField, bodyBytes); err != nil {
					HTTPException(validErrorCode, err.Error())
				}
			} else {
				HTTPException(validErrorCode, h.lang.Required("body"))
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
			childField.MethodByName(field.inType).Call([]reflect.Value{reflect.ValueOf(token)})
		case inTypeSecurityHTTPBasic:
			username, password, _ := req.BasicAuth()
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			childField.MethodByName(field.inType).Call([]reflect.Value{reflect.ValueOf(username), reflect.ValueOf(password)})
		case inTypeSecurityApiKey:
			if !securityApiKey.IsValid() {
				securityApiKey = h.getChildFieldVal(inputValue, field.deepIdx[:len(field.deepIdx)-1])
			}
			childField := h.getChildFieldVal(inputValue, field.deepIdx)
			h.initPtr(childField)
			switch field.inTypeSecurity {
			case inTypeHeader:
				value := req.Header.Get(field.inTypeVal)
				var values []string
				if value != "" {
					fType := field._type
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
				if err := h.handleValue(inputValue, field, values); err != nil {
					HTTPException(validErrorCode, err.Error())
				}
			case inTypeCookie:
				cookie, err := req.Cookie(field.inTypeVal)
				if field._type == typeCookie {
					name := field.tag.desc
					if name == "" {
						name = field.inTypeVal
					}
					if err != nil || cookie.Value == "" {
						if field.mediaTypes[0].required {
							HTTPException(validErrorCode, h.lang.Required(name))
						}
						return
					}
					if err = h.validString(cookie.Value, name, field.tag); err != nil {
						HTTPException(validErrorCode, err.Error())
					}
				} else {
					var values []string
					if cookie.Value != "" {
						fType := field._type
						for fType.Kind() == reflect.Ptr {
							fType = fType.Elem()
						}
						values = []string{cookie.Value}
						if fType.Kind() == reflect.Slice {
							values = []string{}
							valList := strings.Split(cookie.Value, ",")
							for _, v := range valList {
								values = append(values, strings.TrimSpace(v))
							}
						}
					}
					if err = h.handleValue(inputValue, field, values); err != nil {
						HTTPException(validErrorCode, err.Error())
					}
				}
			case inTypeQuery:
				values := req.URL.Query()[field.inTypeVal]
				if err := h.handleValue(inputValue, field, values); err != nil {
					HTTPException(validErrorCode, err.Error())
				}
			}
		}
	}
	if securityApiKey.IsValid() {
		securityApiKey.MethodByName(inTypeSecurityApiKey).Call(nil)
	}
	return
}

func (h *handlerServer) getPaths(path, urlPath string) (rs map[string]string, err error) {
	rs = map[string]string{}
	pathList := strings.Split(path, "/")
	relPathList := strings.Split(urlPath, "/")
	if len(pathList) != len(relPathList) {
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

func (h *handlerServer) handleValue(inputValue reflect.Value, field fieldInfo, values []string) (err error) {
	name := field.tag.desc
	if name == "" {
		name = field.inTypeVal
	}
	required := field.mediaTypes[0].required
	if values == nil {
		if required {
			err = fmt.Errorf(h.lang.Required(name))
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
				err = fmt.Errorf(h.lang.Required(name))
			}
			return
		}
	} else {
		if len(values) == 0 || values[0] == "" {
			if required {
				err = fmt.Errorf(h.lang.Required(name))
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
					err = fmt.Errorf(h.lang.Unique(name))
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
		err = fmt.Errorf(h.lang.Lt(name, *tag.lt))
		return
	}
	if tag.lte != nil && f > *tag.lte {
		err = fmt.Errorf(h.lang.Lte(name, *tag.lte))
		return
	}
	if tag.gt != nil && f <= *tag.gt {
		err = fmt.Errorf(h.lang.Gt(name, *tag.gt))
		return
	}
	if tag.gte != nil && f < *tag.gte {
		err = fmt.Errorf(h.lang.Gte(name, *tag.gte))
		return
	}
	if tag.multiple != nil {
		if *tag.multiple == 0 {
			err = fmt.Errorf(h.lang.MultipleOf(name, *tag.multiple))
			return
		}
		rs, _ := decimal.NewFromFloat(f).Div(decimal.NewFromFloat(*tag.multiple)).Float64()
		if rs != float64(int64(rs)) {
			err = fmt.Errorf(h.lang.MultipleOf(name, *tag.multiple))
			return
		}
	}
	var enum []float64
	for _, v := range tag.enum {
		enum = append(enum, v.(float64))
	}
	if len(enum) > 0 && !inArray(f, enum) {
		err = fmt.Errorf(h.lang.Enum(name, tag.enum))
		return
	}
	return
}

func (h *handlerServer) validString(s string, name string, tag *fieldTagInfo) (err error) {
	if err = h.validLen(len(s), name, tag); err != nil {
		return
	}
	if tag.regexp != "" && !regexp.MustCompile(tag.regexp).MatchString(s) {
		err = fmt.Errorf(h.lang.Regexp(name, tag.regexp))
		return
	}
	var enum []string
	for _, v := range tag.enum {
		enum = append(enum, v.(string))
	}
	if len(enum) > 0 && !inArray(s, enum) {
		err = fmt.Errorf(h.lang.Enum(name, tag.enum))
		return
	}
	return
}

func (h *handlerServer) validLen(l int, name string, tag *fieldTagInfo) (err error) {
	if tag.min > 0 && uint64(l) < tag.min {
		err = fmt.Errorf(h.lang.Min(name, tag.min))
		return
	}
	if tag.max != nil && uint64(l) > *tag.max {
		err = fmt.Errorf(h.lang.Max(name, *tag.max))
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
		sInfo := h.structs[key]
		for _, field := range sInfo.fields {
			name := field.tag.desc
			if name == "" {
				name = field.name
			}
			var mType mediaTypeInfo
			for _, v := range field.mediaTypes {
				if v._type == mediaType {
					mType = v
					break
				}
			}
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
					if mType.required {
						err = fmt.Errorf(h.lang.Required(name))
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
					if mType.required {
						err = fmt.Errorf(h.lang.Required(name))
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
					if mType.required {
						err = fmt.Errorf(h.lang.Required(name))
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
					if mType.required {
						err = fmt.Errorf(h.lang.Required(name))
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
					if mType.required {
						err = fmt.Errorf(h.lang.Required(name))
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
							err = fmt.Errorf(h.lang.Unique(name))
							return
						}
					}
				}
			case reflect.Map:
				vLen := len(v.MapKeys())
				if vLen == 0 {
					if mType.required {
						err = fmt.Errorf(h.lang.Required(name))
						return
					}
					continue
				}
				if err = h.validLen(vLen, name, field.tag); err != nil {
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
