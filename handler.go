package goapi

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func newHandler(api *API) *handler {
	return &handler{api: api}
}

type handler struct {
	api          *API
	paths        []pathInfo
	structFields []fieldInfo
	structs      map[string]*structInfo
	middlewares  []Middleware
}

func (h *handler) Handle() {
	h.middlewares = append(h.middlewares, setLogger())
	for _, hd := range h.api.handlers {
		switch val := hd.(type) {
		case *includeRouter:
			pathMiddlewares := append(h.middlewares, val.middlewares...)
			list := h.handleIncludeRouter(val)
			for k, v := range list {
				v.middlewares = pathMiddlewares
				list[k] = v
			}
			h.paths = append(h.paths, list...)
		case Middleware:
			h.middlewares = append(h.middlewares, val)
		}
	}
	if h.api.httpExceptionResponse != nil {
		resp := fieldInfo{
			deepTypes: h.parseType(reflect.TypeOf(h.api.httpExceptionResponse.GetBody())),
		}
		lastType := resp.deepTypes[len(resp.deepTypes)-1]
		name := ""
		if lastType.isStruct && len(resp.deepTypes) == 1 {
			name = lastType._type.Name()
		}
		for _, mediaType := range h.api.responseMediaTypes {
			resp.mediaTypes = append(resp.mediaTypes, mediaTypeInfo{
				name:     name,
				_type:    mediaTypeToTypeMap[mediaType],
				required: true,
			})
		}
		if lastType.isStruct {
			h.structFields = append(h.structFields, fieldInfo{
				_type:      lastType._type,
				mediaTypes: resp.mediaTypes,
			})
		}
		for k, v := range h.paths {
			v.exceptRes = &resp
			v.respMediaTypes = h.api.responseMediaTypes
			h.paths[k] = v
		}
	}
	h.handleStructs()
}

func (h *handler) handleStructs() {
	h.structs = map[string]*structInfo{}
	structFields := h.structFields
	for len(structFields) > 0 {
		var newStructFields []fieldInfo
		for _, val := range structFields {
			types := h.parseType(val._type)
			lastType := types[len(types)-1]
			if !lastType.isStruct {
				continue
			}
			sType := lastType._type
			stInfo := &structInfo{
				name:  sType.Name(),
				pkg:   sType.PkgPath(),
				_type: sType,
			}
			key := fmt.Sprintf("%v.%v", stInfo.pkg, stInfo.name)
			if key == "." {
				key = fmt.Sprintf("%v%p", prefixTempStruct, sType)
			}
			numField := sType.NumField()
			oldStruct := h.structs[key]
			idx := 0
			for i := 0; i < numField; i++ {
				field := sType.Field(i)
				if field.Name[0] < 'A' || field.Name[0] > 'Z' {
					continue
				}
				childTypes := h.parseType(field.Type)
				lastChildType := childTypes[len(childTypes)-1]
				fFile := fieldInfo{
					name:      field.Name,
					_type:     field.Type,
					deepTypes: childTypes,
					deepIdx:   []int{i},
				}
				tag := field.Tag
				// mediaType
				var mTypes []mediaTypeInfo
				oldTypeMap := map[string]bool{}
				if oldStruct != nil {
					oldTypes := oldStruct.fields[idx].mediaTypes
					for _, v := range oldTypes {
						oldTypeMap[v._type] = true
					}
					mTypes = oldTypes
				}
				for _, v := range val.mediaTypes {
					if oldTypeMap[v._type] {
						continue
					}
					mType := mediaTypeInfo{
						_type:    v._type,
						name:     field.Name,
						required: true,
					}
					tagVal := tag.Get(v._type)
					if tagVal != "" {
						tagList := strings.Split(tagVal, ",")
						mType.name = tagList[0]
						if len(tagList) > 1 && tagList[1] == omitempty {
							mType.required = false
						}
					}
					mTypes = append(mTypes, mType)
				}

				fFile.mediaTypes = mTypes
				// tag
				fTag := &fieldTagInfo{
					regexp: tag.Get(tagRegexp),
					desc:   tag.Get(tagDesc),
				}
				if tagVal := field.Tag.Get(tagEnum); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.enum, field.Type.Kind())
				}
				if tagVal := tag.Get(tagDefault); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag._default, field.Type.Kind())
				}
				if tagVal := tag.Get(tagExample); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.example, field.Type.Kind())
				}
				if tagVal := tag.Get(tagLt); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.lt, field.Type.Kind())
				}
				if tagVal := tag.Get(tagLte); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.lte, field.Type.Kind())
				}
				if tagVal := tag.Get(tagGt); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.gt, field.Type.Kind())
				}
				if tagVal := tag.Get(tagGte); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.gte, field.Type.Kind())
				}
				if tagVal := tag.Get(tagMultiple); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.multiple, field.Type.Kind())
				}
				if tagVal := tag.Get(tagMax); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.max, field.Type.Kind())
				}
				if tagVal := tag.Get(tagMin); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.min, field.Type.Kind())
				}
				if tagVal := tag.Get(tagUnique); tagVal != "" {
					h.parseTagValByKind(tagVal, &fTag.unique, field.Type.Kind())
				}
				fFile.tag = fTag
				if lastChildType.isStruct {
					csType := lastChildType._type
					fFile._struct = &structInfo{
						name:  csType.Name(),
						pkg:   csType.PkgPath(),
						_type: csType,
					}
					newStructFields = append(newStructFields, fieldInfo{
						_type:      csType,
						mediaTypes: val.mediaTypes,
					})
				}
				stInfo.fields = append(stInfo.fields, fFile)
				idx++
			}
			h.structs[key] = stInfo
		}
		structFields = []fieldInfo{}
		for _, val := range newStructFields {
			key := fmt.Sprintf("%v.%v", val._type.PkgPath(), val._type.Name())
			if h.structs[key] == nil {
				structFields = append(structFields, val)
			}
		}
	}
}

func (h *handler) handleIncludeRouter(router *includeRouter) (list []pathInfo) {
	routerType := reflect.ValueOf(router.router)
	if routerType.Kind() != reflect.Ptr || routerType.Elem().Kind() != reflect.Struct {
		panic("Router must be a struct pointer")
	}
	numMethod := routerType.NumMethod()
	for i := 0; i < numMethod; i++ {
		method := routerType.Method(i)
		pInfo := pathInfo{
			funcValue: method,
			isDocs:    router.isDocs,
		}
		numIn := method.Type().NumIn()
		var params []any
		switch numIn {
		case 1:
			if method.Type().In(0).Kind() != reflect.Struct {
				panic("When the method parameter in the router must be 1, it must be a structure")
			}
			params = h.handleInType(routerType.Method(i).Type().In(0), "in", nil)
			pInfo.inTypes = []reflect.Type{
				routerType.Method(i).Type().In(0),
			}
		case 2:
			if method.Type().In(0) != reflect.TypeOf(&Context{}) {
				panic("When the method parameter in the router must be 2, the 1st parameter must be '*goapi.Context'")
			}
			if method.Type().In(1).Kind() != reflect.Struct {
				panic("When the method parameter in the router must be 2, the 2st parameter must be a structure")
			}
			params = h.handleInType(method.Type().In(1), "in", nil)
			pInfo.inTypes = []reflect.Type{
				routerType.Method(i).Type().In(0),
				routerType.Method(i).Type().In(1),
			}
		default:
			panic("The method parameters in the router must be 1 or 2")
		}
		numOut := method.Type().NumOut()
		if numOut != 1 {
			panic("The method parameters out the router must be 1")
		}
		var rInfo *routerInfo
		var fInfoList []fieldInfo
		for _, param := range params {
			switch val := param.(type) {
			case fieldInfo:
				fInfoList = append(fInfoList, val)
			case *routerInfo:
				if rInfo != nil {
					panic("Only one router can exist in the parameters")
				}
				rInfo = val
			}
		}
		respType := routerType.Method(i).Type().Out(0)
		if respType.Implements(typeResponse) {
			resp := reflect.New(respType.Elem()).Interface().(Response)
			respType = reflect.TypeOf(resp.GetBody())
		}
		resp := fieldInfo{
			_type:     respType,
			deepTypes: h.parseType(respType),
		}
		lastType := resp.deepTypes[len(resp.deepTypes)-1]
		name := ""
		if lastType.isStruct && len(resp.deepTypes) == 1 {
			name = lastType._type.Name()
		}
		for _, mediaType := range h.api.responseMediaTypes {
			resp.mediaTypes = append(resp.mediaTypes, mediaTypeInfo{
				name:     name,
				_type:    mediaTypeToTypeMap[mediaType],
				required: true,
			})
		}
		if lastType.isStruct {
			h.structFields = append(h.structFields, fieldInfo{
				_type:      lastType._type,
				mediaTypes: resp.mediaTypes,
			})
		}
		pInfo.path = router.prefix + rInfo.path
		pInfo.methods = rInfo.methods
		pInfo.inputFields = fInfoList
		pInfo.summary = rInfo.summary
		pInfo.desc = rInfo.desc
		pInfo.tags = rInfo.tags
		pInfo.res = &resp
		list = append(list, pInfo)
	}
	return
}

func (h *handler) handleInType(inType reflect.Type, pType string, deepIdx []int) (list []any) {
	numField := inType.NumField()
	for i := 0; i < numField; i++ {
		field := inType.Field(i)
		fType := field.Type
		switch pType {
		case "in":
			switch fType {
			case reflect.TypeOf(Router{}):
				path := field.Tag.Get("path")
				method := field.Tag.Get("method")
				if path == "" || method == "" {
					panic("The parameters must have a path and method present")
				}
				methods := strings.Split(method, ",")
				for k, v := range methods {
					methods[k] = strings.ToUpper(v)
				}
				if !h.isMethod(methods) {
					panic("The method in the parameter does not exist " + strings.Join(methods, ", "))
				}
				summary := field.Tag.Get("summary")
				desc := field.Tag.Get("desc")
				tag := field.Tag.Get("tags")
				var tags []string
				if tag != "" {
					tags = strings.Split(tag, ",")
				}
				list = append(list, &routerInfo{
					path:    path,
					methods: methods,
					summary: summary,
					desc:    desc,
					tags:    tags,
				})
			default:
				if field.Name[0] < 'A' || field.Name[0] > 'Z' {
					continue
				}
				requestType := ""
				for _, inTypeStr := range inTypes {
					tag := field.Tag
					val := tag.Get(inTypeStr)
					if val == "" {
						continue
					}
					if requestType != "" {
						panic(fmt.Sprintf("Field %s cannot have both '%s' and '%s' labels present at the same time",
							field.Name, requestType, inTypeStr))
					}
					valList := strings.Split(val, ",")
					required := true
					if len(valList) > 1 && valList[1] == omitempty {
						required = false
					}
					fInfo := fieldInfo{
						name:       field.Name,
						_type:      fType,
						inType:     inTypeStr,
						inTypeVal:  valList[0],
						mediaTypes: []mediaTypeInfo{{required: required}},
						deepIdx:    append(deepIdx, i),
					}
					// tag
					fTag := &fieldTagInfo{
						regexp: tag.Get(tagRegexp),
						desc:   tag.Get(tagDesc),
					}
					if tagVal := field.Tag.Get(tagEnum); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.enum, field.Type.Kind())
					}
					if tagVal := tag.Get(tagDefault); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag._default, field.Type.Kind())
					}
					if tagVal := tag.Get(tagExample); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.example, field.Type.Kind())
					}
					if tagVal := tag.Get(tagLt); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.lt, field.Type.Kind())
					}
					if tagVal := tag.Get(tagLte); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.lte, field.Type.Kind())
					}
					if tagVal := tag.Get(tagGt); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.gt, field.Type.Kind())
					}
					if tagVal := tag.Get(tagGte); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.gte, field.Type.Kind())
					}
					if tagVal := tag.Get(tagMultiple); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.multiple, field.Type.Kind())
					}
					if tagVal := tag.Get(tagMax); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.max, field.Type.Kind())
					}
					if tagVal := tag.Get(tagMin); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.min, field.Type.Kind())
					}
					if tagVal := tag.Get(tagUnique); tagVal != "" {
						h.parseTagValByKind(tagVal, &fTag.unique, field.Type.Kind())
					}
					requestType = inTypeStr
					switch inTypeStr {
					case inTypeHeader, inTypeCookie, inTypeQuery, inTypeForm, inTypePath:
						for fType.Kind() == reflect.Ptr {
							if fType == typeCookie && inTypeStr == inTypeCookie {
								break
							}
							fType = fType.Elem()
						}
						fInfo.deepTypes = h.parseType(fType)
						if inTypeStr != inTypeCookie || fType != typeCookie {
							if fType.Kind() == reflect.Slice {
								fType = fType.Elem()
							}
							if !h.isNormalType(fType) {
								panic(fmt.Sprintf("The %s type must be a regular type", inTypeStr))
							}
						}
						fInfo.tag = fTag
					case inTypeBody:
						var mTypes []mediaTypeInfo
						for _, v := range valList {
							if v != jsonType && v != xmlType {
								panic("The body must in 'json','xml'")
							}
							mType := mediaTypeInfo{
								_type:    v,
								name:     field.Name,
								required: true,
							}
							childVal := tag.Get(v)
							if childVal != "" {
								childValList := strings.Split(childVal, ",")
								mType.name = childValList[0]
								if len(childValList) > 1 && childValList[1] == omitempty {
									mType.required = false
								}
							}
							mTypes = append(mTypes, mType)
						}
						fInfo.mediaTypes = mTypes
						fInfo.deepTypes = h.parseType(fType)
						fInfo.inType = inTypeStr
						lastType := fInfo.deepTypes[len(fInfo.deepTypes)-1]
						if lastType.isStruct {
							h.structFields = append(h.structFields, fieldInfo{
								_type:      lastType._type,
								mediaTypes: mTypes,
							})
						}
					case inTypeFile:
						if fType != typeFile && fType != typeFiles {
							panic("The type of file must in '*multipart.FileHeader', '[]*multipart.FileHeader")
						}
						fInfo.deepTypes = []typeInfo{{_type: fType}}
						if fType == typeFiles {
							fInfo.deepTypes = append(fInfo.deepTypes, typeInfo{_type: fType.Elem()})
						}
						fInfo.tag = fTag
					}
					list = append(list, fInfo)
				}
				if requestType != "" {
					continue
				}
				if fType.Kind() == reflect.Ptr {
					if securityList, ok := h.handleSecurity(fType, field.Name, append(deepIdx, i)); ok {
						for _, security := range securityList {
							list = append(list, security)
						}
						continue
					}
				}
				if isFixedType(fType) {
					continue
				}
				for fType.Kind() == reflect.Ptr {
					fType = fType.Elem()
				}
				if fType.Kind() == reflect.Struct {
					list = append(list, h.handleInType(fType, pType, append(deepIdx, i))...)
				}
			}
		case "apiKey":
			if field.Name[0] < 'A' || field.Name[0] > 'Z' {
				continue
			}
			requestType := ""
			for _, inTypeStr := range []string{inTypeHeader, inTypeCookie, inTypeQuery} {
				val := field.Tag.Get(inTypeStr)
				if val == "" {
					continue
				}
				if requestType != "" {
					panic(fmt.Sprintf("Field %s cannot have both '%s' and '%s' labels present at the same time",
						field.Name, requestType, inTypeStr))
				}
				fInfo := fieldInfo{
					name:      field.Name,
					_type:     fType,
					inType:    inTypeStr,
					inTypeVal: val,
					deepIdx:   append(deepIdx, i),
				}
				requestType = inTypeStr
				for fType.Kind() == reflect.Ptr {
					fType = fType.Elem()
				}
				if fType.Kind() == reflect.Slice {
					fType = fType.Elem()
				}
				if !h.isNormalType(fType) {
					panic(fmt.Sprintf("The %s type must be a regular type", inTypeStr))
				}
				list = append(list, fInfo)
			}
			if requestType != "" {
				continue
			}
			for fType.Kind() == reflect.Ptr {
				fType = fType.Elem()
			}
			if fType.Kind() == reflect.Struct {
				list = append(list, h.handleInType(fType, pType, append(deepIdx, i))...)
			}
		}
	}
	return
}

func (h *handler) isNormalType(fType reflect.Type) bool {
	switch fType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool, reflect.String:
	default:
		return false
	}
	return true
}

func (h *handler) isMethod(methods []string) bool {
	for _, method := range methods {
		switch method {
		case http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions, http.MethodHead,
			http.MethodPatch, http.MethodTrace:
		default:
			return false
		}
	}
	return true
}

func (h *handler) parseType(fType reflect.Type) (rs []typeInfo) {
	for fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}
	switch fType.Kind() {
	case reflect.Map, reflect.Slice:
		rs = append([]typeInfo{{_type: fType}}, h.parseType(fType.Elem())...)
	case reflect.Struct:
		rs = append(rs, typeInfo{_type: fType, isStruct: true})
	default:
		rs = append(rs, typeInfo{_type: fType})
	}
	return
}

func (h *handler) handleSecurity(fType reflect.Type, fName string, deepIdx []int) (list []fieldInfo, ok bool) {
	num := 0
	for _, securityType := range securityTypes {
		if !fType.Implements(securityType) {
			continue
		}
		if num > 0 {
			panic("Security can only implement one of the interfaces 'HTTPBearerInterface', " +
				"'HTTPBasicInterface', and 'ApiKeyInterface'")
		}
		num++
		switch securityType {
		case securityTypeHTTPBearer:
			list = append(list, fieldInfo{
				name:    fType.Elem().Name(),
				_type:   fType,
				inType:  inTypeSecurityHTTPBearer,
				deepIdx: deepIdx,
			})
		case securityTypeHTTPBasic:
			list = append(list, fieldInfo{
				name:    fType.Elem().Name(),
				_type:   fType,
				inType:  inTypeSecurityHTTPBasic,
				deepIdx: deepIdx,
			})
		case securityTypeApiKey:
			cList := h.handleInType(fType.Elem(), "apiKey", deepIdx)
			for _, v := range cList {
				if f, b := v.(fieldInfo); b {
					f.inTypeSecurity = f.inType
					f.inType = inTypeSecurityApiKey
					list = append(list, f)
				}
			}
		}
	}
	ok = num > 0
	return
}

func (h *handler) parseTagValByKind(inVal string, outVal any, kind reflect.Kind) {
	switch val := outVal.(type) {
	case *string:
		*val = inVal
	case *float64:
		if v, err := strconv.ParseFloat(inVal, 64); err == nil {
			*val = v
		} else {
			panic(err)
		}
	case **float64:
		if v, err := strconv.ParseFloat(inVal, 64); err == nil {
			*val = toPtr(v)
		} else {
			panic(err)
		}
	case *uint64:
		if v, err := strconv.ParseUint(inVal, 10, 64); err == nil {
			*val = v
		} else {
			panic(err)
		}
	case **uint64:
		if v, err := strconv.ParseUint(inVal, 10, 64); err == nil {
			*val = toPtr(v)
		} else {
			panic(err)
		}
	case *bool:
		if v, err := strconv.ParseBool(inVal); err == nil {
			*val = v
		} else {
			panic(err)
		}
	case *any:
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
			reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			if v, err := strconv.ParseFloat(inVal, 64); err == nil {
				*val = toPtr(v)
			} else {
				panic(err)
			}
		case reflect.Bool:
			if v, err := strconv.ParseBool(inVal); err == nil {
				*val = v
			} else {
				panic(err)
			}
		case reflect.String:
			*val = inVal
		default:
			panic("Structure tag value type error")
		}
	case *[]any:
		list := strings.Split(inVal, ",")
		var rs []any
		for _, str := range list {
			str = strings.TrimSpace(str)
			switch kind {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
				reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				if v, err := strconv.ParseFloat(str, 64); err == nil {
					rs = append(rs, v)
				} else {
					panic(err)
				}
			case reflect.Bool:
				if v, err := strconv.ParseBool(str); err == nil {
					rs = append(rs, v)
				} else {
					panic(err)
				}
			case reflect.String:
				rs = append(rs, str)
			default:
				panic("Structure tag value type error")
			}
		}
		*val = rs
	}
}
