package goapi

import (
	"fmt"
	"github.com/goodluckxu-go/goapi/openapi"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func newHandler(api *API) *handler {
	return &handler{
		api:            api,
		allMediaTypes:  map[MediaType]struct{}{},
		openapiSetMap:  map[string]*openapi.OpenAPI{},
		childPrefixMap: map[string]struct{}{},
		sameRoute:      map[string]struct{}{},
	}
}

type handler struct {
	api                *API
	paths              []*pathInfo
	statics            []*staticInfo
	structFields       []fieldInfo
	structs            map[string]*structInfo
	defaultMiddlewares []Middleware
	publicMiddlewares  []Middleware
	allMediaTypes      map[MediaType]struct{}
	openapiSetMap      map[string]*openapi.OpenAPI
	childPrefixMap     map[string]struct{}
	sameRoute          map[string]struct{}
}

func (h *handler) Handle() {
	for _, v := range h.api.responseMediaTypes {
		h.allMediaTypes[v] = struct{}{}
	}
	h.defaultMiddlewares = append(h.defaultMiddlewares, setLogger())
	h.publicMiddlewares = h.handleHandlers(h.api.handlers, h.defaultMiddlewares, "", true, h.api.docsPath)
	h.publicMiddlewares = append(h.defaultMiddlewares, h.publicMiddlewares...)
	if h.api.httpExceptionResponse != nil {
		resp := fieldInfo{
			deepTypes: h.parseType(reflect.TypeOf(h.api.httpExceptionResponse.GetBody())),
		}
		lastType := resp.deepTypes[len(resp.deepTypes)-1]
		resp.mediaTypes = h.api.responseMediaTypes
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
	if err := h.handleStructs(); err != nil {
		log.Fatal(err)
	}
}

func (h *handler) handleHandlers(handlers []any, middlewares []Middleware, prefix string, isDocs bool, docsPath string) (public []Middleware) {
	for _, hd := range handlers {
		switch val := hd.(type) {
		case *includeRouter:
			pathMiddlewares := append(middlewares, val.middlewares...)
			list, err := h.handleIncludeRouter(val, prefix)
			if err != nil {
				log.Fatal(err)
			}
			for k, v := range list {
				v.middlewares = pathMiddlewares
				v.isDocs = isDocs && v.isDocs
				v.docsPath = docsPath
				list[k] = v
				for _, v1 := range v.methods {
					tmpRoute := fmt.Sprintf("%v_%v", v1, v.path)
					if _, ok := h.sameRoute[tmpRoute]; ok {
						log.Fatal(fmt.Sprintf("there are multiple methods for '%v' and routing '%v'", v1, v.path))
					}
					h.sameRoute[tmpRoute] = struct{}{}
				}
			}
			h.paths = append(h.paths, list...)
		case *staticInfo:
			h.statics = append(h.statics, val)
		case *APIGroup:
			h.handleHandlers(val.handlers, middlewares, prefix+val.prefix, isDocs && val.isDocs, docsPath)
		case *ChildAPI:
			if val.docsPath == "" {
				log.Fatal("childAPI must have docsPath")
			}
			if h.openapiSetMap[docsPath+val.docsPath] != nil {
				log.Fatal("the childAPI docsPath repeats")
			}
			if val.prefix == "" {
				log.Fatal("childAPI must have prefix")
			}
			if _, ok := h.childPrefixMap[prefix+val.prefix]; ok {
				log.Fatal("the childAPI prefix repeats")
			}
			h.childPrefixMap[prefix+val.prefix] = struct{}{}
			if val.isDocs {
				h.openapiSetMap[docsPath+val.docsPath] = &openapi.OpenAPI{
					Info:    val.OpenAPIInfo,
					Servers: val.OpenAPIServers,
					Tags:    val.OpenAPITags,
				}
			}
			h.handleHandlers(val.handlers, middlewares, prefix+val.prefix, isDocs && val.isDocs, docsPath+val.docsPath)
		case Middleware:
			middlewares = append(middlewares, val)
			public = append(public, val)
		}
	}
	return
}

func (h *handler) handleStructs() (err error) {
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
			oldStruct := h.structs[key]
			if oldStruct != nil {
				continue
			}
			numField := sType.NumField()
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
					fieldMap:  map[MediaType]*fieldNameInfo{},
				}
				tag := field.Tag
				// fieldNameInfo
				for _, v := range bodyMediaTypes {
					fInfo := &fieldNameInfo{
						name:     field.Name,
						required: true,
					}
					tagVal := tag.Get(mediaTypeToTypeMap[v])
					if tagVal != "" {
						tagList := strings.Split(tagVal, ",")
						if tagList[0] != "" {
							nameList := strings.Split(tagList[0], ">")
							if len(nameList) == 1 {
								fInfo.name = tagList[0]
							} else {
								fInfo.name = nameList[0]
								fInfo.xml = &xmlInfo{
									childs: nameList[1:],
								}
							}
						}
						for _, tv := range tagList[1:] {
							switch tv {
							case omitempty:
								fInfo.required = false
							case "attr":
								if v == XML {
									fInfo.xml = &xmlInfo{
										attr: true,
									}
								}
							case "innerxml", "chardata":
								if v == XML {
									fInfo.xml = &xmlInfo{
										innerxml: true,
									}
								}
							}
						}
					}
					fFile.fieldMap[v] = fInfo
				}
				// tag
				fTag := &fieldTagInfo{}
				if fTag, err = h.handleTag(tag, field.Type); err != nil {
					return
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
						_type: csType,
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
			oldStruct := h.structs[key]
			if oldStruct == nil {
				structFields = append(structFields, val)
			}
		}
	}
	return
}

func (h *handler) handleIncludeRouter(router *includeRouter, prefix string) (list []*pathInfo, err error) {
	routerType := reflect.ValueOf(router.router)
	var routerStructType reflect.Type
	var pos string
	switch routerType.Kind() {
	case reflect.Ptr:
		if routerType.Elem().Kind() != reflect.Struct {
			err = fmt.Errorf("router must be a struct or struct pointer")
		}
		routerStructType = routerType.Elem().Type()
		pos = fmt.Sprintf("%v.(*%v)", routerStructType.PkgPath(), routerStructType.Name())
	case reflect.Struct:
		routerStructType = routerType.Type()
		pos = fmt.Sprintf("%v.%v", routerStructType.PkgPath(), routerStructType.Name())
	default:
		err = fmt.Errorf("router must be a struct or struct pointer")
		return
	}
	numMethod := routerType.NumMethod()
	for i := 0; i < numMethod; i++ {
		method := routerType.Method(i)
		pInfo := pathInfo{
			funcValue: method,
			isDocs:    router.isDocs,
			pos:       fmt.Sprintf("%v.%v", pos, routerType.Type().Method(i).Name),
		}
		numIn := method.Type().NumIn()
		var params []any
		switch numIn {
		case 1:
			if method.Type().In(0).Kind() != reflect.Struct {
				err = fmt.Errorf("when the method parameter in the router must be 1, it must be a structure")
				return
			}
			params, err = h.handleInType(routerType.Method(i).Type().In(0), "in", nil)
			if err != nil {
				return
			}
			pInfo.inTypes = []reflect.Type{
				routerType.Method(i).Type().In(0),
			}
		case 2:
			if method.Type().In(0) != reflect.TypeOf(&Context{}) {
				err = fmt.Errorf("when the method parameter in the router must be 2, the 1st parameter must be '*goapi.Context'")
				return
			}
			if method.Type().In(1).Kind() != reflect.Struct {
				err = fmt.Errorf("when the method parameter in the router must be 2, the 2st parameter must be a structure")
				return
			}
			params, err = h.handleInType(method.Type().In(1), "in", nil)
			if err != nil {
				return
			}
			pInfo.inTypes = []reflect.Type{
				routerType.Method(i).Type().In(0),
				routerType.Method(i).Type().In(1),
			}
		default:
			err = fmt.Errorf("the method parameters in the router must be 1 or 2")
			return
		}
		var rInfo *routerInfo
		var fInfoList []fieldInfo
		for _, param := range params {
			switch val := param.(type) {
			case fieldInfo:
				fInfoList = append(fInfoList, val)
			case *routerInfo:
				if rInfo != nil {
					err = fmt.Errorf("only one router can exist in the parameters")
					return
				}
				rInfo = val
			}
		}
		if rInfo == nil {
			err = fmt.Errorf("a route must exist in the parameters")
			return
		}
		var resp *fieldInfo
		if method.Type().NumOut() == 1 {
			var mediaTypes []MediaType
			respType := routerType.Method(i).Type().Out(0)
			if respType.Implements(typeResponse) {
				res := reflect.New(respType.Elem()).Interface().(Response)
				respType = reflect.TypeOf(res.GetBody())
				if res.GetContentType() != "" {
					mediaTypes = append(mediaTypes, MediaType(res.GetContentType()))
				}
			}
			resp = &fieldInfo{
				_type:      respType,
				deepTypes:  h.parseType(respType),
				mediaTypes: mediaTypes,
			}
			lastType := resp.deepTypes[len(resp.deepTypes)-1]
			if len(resp.mediaTypes) == 0 {
				resp.mediaTypes = h.api.responseMediaTypes
			}
			if lastType.isStruct {
				h.structFields = append(h.structFields, fieldInfo{
					_type:      lastType._type,
					mediaTypes: resp.mediaTypes,
				})
			}
		}
		pInfo.path = prefix + router.prefix + rInfo.path
		pInfo.methods = rInfo.methods
		pInfo.inputFields = fInfoList
		pInfo.summary = rInfo.summary
		pInfo.desc = rInfo.desc
		pInfo.tags = rInfo.tags
		if h.api.autoTagsIndex != nil && len(pInfo.tags) == 0 {
			pList := strings.Split(strings.TrimPrefix(pInfo.path, "/"), "/")
			if len(pList) <= *h.api.autoTagsIndex {
				err = fmt.Errorf("the 'index' in method 'SetAutoTags' exceeds the limit")
				return
			}
			pInfo.tags = []string{pList[*h.api.autoTagsIndex]}
		}
		pInfo.res = resp
		list = append(list, &pInfo)
	}
	h.handlePathSort(list)
	return
}

func (h *handler) handlePathSort(list []*pathInfo) {
	sort.Slice(list, func(i, j int) bool {
		return list[i].path < list[j].path
	})
}

func (h *handler) handleInType(inType reflect.Type, pType string, deepIdx []int) (list []any, err error) {
	numField := inType.NumField()
	for i := 0; i < numField; i++ {
		field := inType.Field(i)
		fType := field.Type
		switch pType {
		case "in":
			switch fType {
			case reflect.TypeOf(Router{}):
				path := field.Tag.Get(tagPath)
				method := field.Tag.Get(tagMethod)
				if path == "" || method == "" {
					err = fmt.Errorf("the parameters must have a path and method present")
					return
				}
				methods := strings.Split(method, ",")
				for k, v := range methods {
					methods[k] = strings.ToUpper(v)
				}
				if !h.isMethod(methods) {
					err = fmt.Errorf("the method in the parameter does not exist %v", strings.Join(methods, ", "))
					return
				}
				summary := h.getMappingTag(field.Tag.Get(tagSummary))
				desc := h.getMappingTag(field.Tag.Get(tagDesc))
				tag := field.Tag.Get(tagTags)
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
						err = fmt.Errorf("field %s cannot have both '%s' and '%s' labels present at the same time",
							field.Name, requestType, inTypeStr)
						return
					}
					valList := strings.Split(val, ",")
					required := true
					if len(valList) > 1 && valList[1] == omitempty {
						required = false
					}
					fInfo := fieldInfo{
						name:      field.Name,
						_type:     fType,
						inType:    inTypeStr,
						inTypeVal: valList[0],
						deepIdx:   append(deepIdx, i),
						required:  required,
					}
					// tag
					fTag := &fieldTagInfo{}
					if fTag, err = h.handleTag(tag, field.Type); err != nil {
						return
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
								err = fmt.Errorf("the %s type must be a regular type", inTypeStr)
								return
							}
						}
						if inTypeStr == inTypeCookie {
							fTag.desc += "Read the value of document.cookie"
							fTag.example = "document.cookie"
						}
						fInfo.tag = fTag
					case inTypeBody:
						var mTypes []MediaType
						isNotJsonXml := false
						for _, v := range valList {
							mediaType := typeToMediaTypeMap[v]
							if mediaType == "" {
								mediaType = MediaType(v)
							}
							if mediaType != XML && mediaType != JSON {
								isNotJsonXml = true
							} else {
								h.allMediaTypes[mediaType] = struct{}{}
							}
							mTypes = append(mTypes, mediaType)
						}
						fInfo.mediaTypes = mTypes
						if isNotJsonXml {
							if fType.Kind() != reflect.String && fType != typeBytes && !fType.Implements(interfaceIoReadCloser) {
								err = fmt.Errorf("other media types only support types '[]byte', 'string', and 'io.ReadCloser‘")
								return
							}
							fInfo.deepTypes = []typeInfo{{_type: fType}}
						} else {
							fInfo.deepTypes = h.parseType(fType)
						}
						fInfo.inType = inTypeStr
						lastType := fInfo.deepTypes[len(fInfo.deepTypes)-1]
						if lastType.isStruct {
							h.structFields = append(h.structFields, fieldInfo{
								_type:      lastType._type,
								mediaTypes: mTypes,
							})
						}
						fInfo.tag = fTag
					case inTypeFile:
						if fType != typeFile && fType != typeFiles {
							err = fmt.Errorf("the type of file must in '*multipart.FileHeader', '[]*multipart.FileHeader")
							return
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
					securityList, ok, er := h.handleSecurity(fType, append(deepIdx, i))
					if er != nil {
						err = er
						return
					}
					if ok {
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
					var childList []any
					childList, err = h.handleInType(fType, pType, append(deepIdx, i))
					if err != nil {
						return
					}
					list = append(list, childList...)
				}
			}
		case "apiKey":
			if field.Name[0] < 'A' || field.Name[0] > 'Z' {
				continue
			}
			if field.Type == typeContext {
				continue
			}
			requestType := ""
			for _, inTypeStr := range []string{inTypeHeader, inTypeCookie, inTypeQuery} {
				tag := field.Tag
				val := field.Tag.Get(inTypeStr)
				if val == "" {
					continue
				}
				if requestType != "" {
					err = fmt.Errorf("field %s cannot have both '%s' and '%s' labels present at the same time",
						field.Name, requestType, inTypeStr)
					return
				}
				valList := strings.Split(val, ",")
				required := true
				if len(valList) > 1 && valList[1] == omitempty {
					required = false
				}
				fInfo := fieldInfo{
					name:      field.Name,
					_type:     fType,
					inType:    inTypeStr,
					inTypeVal: valList[0],
					deepIdx:   append(deepIdx, i),
					required:  required,
				}
				// tag
				fTag := &fieldTagInfo{}
				if fTag, err = h.handleTag(tag, field.Type); err != nil {
					return
				}
				if inTypeStr == inTypeCookie {
					fTag.desc += "Read the value of document.cookie"
				}
				fInfo.tag = fTag
				requestType = inTypeStr
				for fType.Kind() == reflect.Ptr {
					fType = fType.Elem()
				}
				if fType.Kind() == reflect.Slice {
					fType = fType.Elem()
				}
				if !h.isNormalType(fType) {
					err = fmt.Errorf("the %s type must be a regular type", inTypeStr)
					return
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
				var childList []any
				childList, err = h.handleInType(fType, pType, append(deepIdx, i))
				if err != nil {
					return
				}
				list = append(list, childList...)
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
	if fType == nil {
		rs = append(rs, typeInfo{_type: typeAny})
		return
	}
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

func (h *handler) handleSecurity(fType reflect.Type, deepIdx []int) (list []fieldInfo, ok bool, err error) {
	num := 0
	for _, securityType := range securityTypes {
		if !fType.Implements(securityType) {
			continue
		}
		if num > 0 {
			err = fmt.Errorf("security can only implement one of the interfaces 'goapi.HTTPBearer', " +
				"'goapi.HTTPBasic', and 'goapi.ApiKey'")
			return
		}
		fTag := &fieldTagInfo{}
		value := reflect.New(fType.Elem())
		var desc SecurityDescription
		if desc, ok = value.Interface().(SecurityDescription); ok {
			fTag.desc = desc.Desc()
		}
		num++
		switch securityType {
		case securityTypeHTTPBearer:
			list = append(list, fieldInfo{
				name:    fType.Elem().Name(),
				_type:   fType,
				inType:  inTypeSecurityHTTPBearer,
				deepIdx: deepIdx,
				tag:     fTag,
			})
		case securityTypeHTTPBearerJWT:
			list = append(list, fieldInfo{
				name:    fType.Elem().Name(),
				_type:   fType,
				inType:  inTypeSecurityHTTPBearerJWT,
				deepIdx: deepIdx,
				tag:     fTag,
			})
		case securityTypeHTTPBasic:
			list = append(list, fieldInfo{
				name:    fType.Elem().Name(),
				_type:   fType,
				inType:  inTypeSecurityHTTPBasic,
				deepIdx: deepIdx,
				tag:     fTag,
			})
		case securityTypeApiKey:
			var cList []any
			cList, err = h.handleInType(fType.Elem(), "apiKey", deepIdx)
			if err != nil {
				return
			}
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

func (h *handler) parseTagValByKind(inVal string, outVal any, kind reflect.Kind) error {
	switch val := outVal.(type) {
	case *string:
		*val = inVal
	case *float64:
		if v, err := strconv.ParseFloat(inVal, 64); err == nil {
			*val = v
		} else {
			return err
		}
	case **float64:
		if v, err := strconv.ParseFloat(inVal, 64); err == nil {
			*val = toPtr(v)
		} else {
			return err
		}
	case *uint64:
		if v, err := strconv.ParseUint(inVal, 10, 64); err == nil {
			*val = v
		} else {
			return err
		}
	case **uint64:
		if v, err := strconv.ParseUint(inVal, 10, 64); err == nil {
			*val = toPtr(v)
		} else {
			return err
		}
	case *bool:
		if v, err := strconv.ParseBool(inVal); err == nil {
			*val = v
		} else {
			return err
		}
	case *any:
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
			reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			if v, err := strconv.ParseFloat(inVal, 64); err == nil {
				*val = toPtr(v)
			} else {
				return err
			}
		case reflect.Bool:
			if v, err := strconv.ParseBool(inVal); err == nil {
				*val = v
			} else {
				return err
			}
		case reflect.String:
			*val = inVal
		default:
			return fmt.Errorf("structure tag value type error")
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
					return err
				}
			case reflect.Bool:
				if v, err := strconv.ParseBool(str); err == nil {
					rs = append(rs, v)
				} else {
					return err
				}
			case reflect.String:
				rs = append(rs, str)
			default:
				return fmt.Errorf("structure tag value type error")
			}
		}
		*val = rs
	}
	return nil
}

func (h *handler) handleTag(tag reflect.StructTag, fType reflect.Type) (fTag *fieldTagInfo, err error) {
	fKind := fType.Kind()
	if fType.Implements(interfaceToStringer) {
		fKind = reflect.String
	}
	fTag = &fieldTagInfo{
		regexp: tag.Get(tagRegexp),
		desc:   h.getMappingTag(tag.Get(tagDesc)),
	}
	if tagVal := tag.Get(tagEnum); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.enum, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagDefault); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag._default, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagExample); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.example, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagDeprecated); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.deprecated, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagLt); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.lt, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagLte); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.lte, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagGt); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.gt, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagGte); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.gte, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagMultiple); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.multiple, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagMax); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.max, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagMin); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.min, fKind); err != nil {
			return
		}
	}
	if tagVal := tag.Get(tagUnique); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &fTag.unique, fKind); err != nil {
			return
		}
	}
	return
}

func (h *handler) getMappingTag(tagVal string, replaces ...map[string]int) string {
	n := len(tagVal)
	left := 0 // 变量左边坐标
	i := 0
	var buf []byte
	isMapping := false
	replace := map[string]int{}
	if len(replaces) > 0 {
		replace = replaces[0]
	}
	replaceNum := map[string]int{}
	for i < n {
		if i+1 < n && tagVal[i:i+2] == "{{" {
			if i+2 < n && tagVal[i+2] == '{' {
				buf = append(buf, tagVal[i])
				i++
				continue
			}
			buf = append(buf, "{{"...)
			left = i + 2
			i += 2
			continue
		} else if left > 0 && i+1 < n && tagVal[i:i+2] == "}}" {
			oldVal := tagVal[left:i]
			val := h.api.structTagVariableMap[oldVal]
			if val != nil {
				buf = append(buf[0:len(buf)-len(oldVal)-2], val.(string)...)
				isMapping = true
				replaceNum[oldVal]++
			} else {
				buf = append(buf, "}}"...)
			}
			left = 0
			i += 2
			continue
		}
		buf = append(buf, tagVal[i])
		i++
	}
	for k, _ := range replaceNum {
		if replace[k] > 0 {
			log.Fatal(fmt.Sprintf("mapping tag '%v' dead loop", k))
		}
		replace[k]++
	}
	if isMapping {
		return h.getMappingTag(string(buf), replace)
	}
	return string(buf)
}
