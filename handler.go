package goapi

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/textproto"
	"reflect"
	"strconv"
	"strings"

	"github.com/goodluckxu-go/goapi/v2/openapi"
	"github.com/goodluckxu-go/goapi/v2/swagger"
)

func newHandler(api *API) *handler {
	return &handler{
		api:                    api,
		paths:                  make([]*pathInfo, 0),
		structDepends:          make(map[string][]string),
		structs:                make(map[string]*structInfo),
		structTypes:            make(map[string]reflect.Type),
		mediaTypes:             map[MediaType]struct{}{},
		publicGroupMiddlewares: make(map[string][]HandleFunc),
		openapiMap:             map[string]*openapi.OpenAPI{},
		swaggerMap:             map[string]swagger.Config{},
		exceptMap:              map[string]*exceptInfo{},
	}
}

type handler struct {
	api                    *API
	paths                  []*pathInfo
	structDepends          map[string][]string
	structs                map[string]*structInfo
	structTypes            map[string]reflect.Type
	mediaTypes             map[MediaType]struct{}
	publicGroupMiddlewares map[string][]HandleFunc // group prefix
	openapiMap             map[string]*openapi.OpenAPI
	swaggerMap             map[string]swagger.Config
	childMap               map[string]returnObjChild
	exceptMap              map[string]*exceptInfo
}

func (h *handler) Handle() {
	if len(h.api.responseMediaTypes) == 0 {
		h.api.responseMediaTypes = []MediaType{JSON}
	}
	obj, err := h.api.returnObj()
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range obj.groupMap {
		if len(v.middlewares) > 0 {
			h.publicGroupMiddlewares[k] = append(h.publicGroupMiddlewares[k], v.middlewares...)
		}
	}
	h.childMap = obj.childMap
	for k, v := range h.childMap {
		h.exceptMap[k] = &exceptInfo{
			exceptFunc: v.exceptFunc,
		}
	}
	for k, v := range obj.docsMap {
		if !v.isDocs {
			continue
		}
		v.info.Summary = h.getMappingTag(v.info.Summary)
		v.info.Description = h.getMappingTag(v.info.Description)
		h.openapiMap[k] = &openapi.OpenAPI{
			Info:    v.info,
			Servers: v.servers,
			Tags:    v.tags,
		}
		h.swaggerMap[k] = v.swagger
	}
	for _, v := range h.api.responseMediaTypes {
		h.mediaTypes[v] = struct{}{}
	}
	for k, v := range obj.mediaTypes {
		h.mediaTypes[k] = v
	}
	h.paths = obj.paths
	var field *paramField
	for _, path := range h.paths {
		if path.inFs != nil {
			continue
		}
		path.desc = h.getMappingTag(path.desc)
		path.summary = h.getMappingTag(path.summary)
		for key, in := range path.inParams {
			if in.parentInType != "" && !inArray(in.inType, []InType{inTypeQuery, inTypeHeader, inTypeCookie}) {
				log.Fatal("only 'query','header' and 'cookie' can be passed into interface security")
			}
			if in.inType == inTypeFile {
				if !isArrayType(in.structField.Type, func(sType reflect.Type) bool {
					if sType.ConvertibleTo(typeFile) {
						return true
					}
					return false
				}, 2) {
					log.Fatal(fmt.Sprintf("the type of parameter '%v' in '%v' must be "+
						"‘*multipart.FileHeader’ or an array of ‘*multipart.FileHeader’, has type '%v'",
						in.values[0].name, in.inType.Tag(),
						in.structField.Type.String()))
				}
			} else if in.inType == inTypeBody {
				m := map[uint8]struct{}{}
				var mList []string
				for _, val := range in.values {
					if val.mediaType.IsStream() {
						// stream
						m[1] = struct{}{}
					} else {
						// no stream
						m[0] = struct{}{}
					}
					mv := "'" + string(val.mediaType) + "'"
					if !inArray(mv, mList) {
						mList = append(mList, mv)
					}
				}
				if len(m) != 1 {
					log.Fatal(fmt.Sprintf("Content-Type %v cannot be used together", strings.Join(mList, ", ")))
				}
				if _, ok := m[1]; ok {
					// stream
					vType := in.structField.Type
					for vType.Kind() == reflect.Ptr {
						vType = vType.Elem()
					}
					if !(vType.ConvertibleTo(typeBytes) || vType.Kind() == reflect.String || vType == typeReadCloser) {
						log.Fatal("other media types only support types '[]byte', 'string', and 'io.ReadCloser‘")
					}
					in.field = &paramField{
						tag: &paramTag{},
					}
					fType := removeMorePtr(in.structField.Type)
					in.field.kind = fType.Kind()
					if fType.Kind() == reflect.Ptr {
						in.field.kind = fType.Elem().Kind()
					}
					fVal := getValueByType(fType, true)
					if err = h.handleTagByInterface(fType, in.field, fVal); err != nil {
						log.Fatal(err)
					}
					if fType.Kind() == reflect.Ptr {
						fType = fType.Elem()
					}
					in.field._type = fType
					if err = h.handleTagByField(in.structField, in.field); err != nil {
						log.Fatal(err)
					}
				} else {
					// no stream
					field, err = h.handleField(in.structField, -1)
					if err != nil {
						log.Fatal(err)
					}
					field.anonymous = true
					in.field = field
				}
				path.inParams[key] = in
				continue
			} else if in.inType.IsSingle() {
				if !isArrayType(in.structField.Type, func(sType reflect.Type) bool {
					if isNormalType(sType) {
						return true
					}
					if in.inType == inTypeCookie && sType.ConvertibleTo(typeCookie) {
						return true
					}
					if isTextInterface(sType) {
						return true
					}
					return false
				}, 2) {
					log.Fatal(fmt.Errorf("the type of parameter '%v' in '%v' cannot be '%v'", in.values[0].name,
						in.inType.Tag(), in.structField.Type.String()))
				}
			}
			field, err = h.handleParam(in.inType, in.structField, -1, in.values)
			if err != nil {
				log.Fatal(err)
			}
			in.field = field
			path.inParams[key] = in
		}
		if path.outParam != nil {
			field = &paramField{
				tag:   &paramTag{},
				_type: path.outParam.structField.Type,
			}
			path.outParam.httpStatus = http.StatusOK
			h.handleOutParam(path.outParam)
			if _, ok := getTypeByCovertInterface[io.ReadCloser](path.outParam.structField.Type); !ok &&
				path.outParam.structField.Type != nil {
				field, err = h.handleField(path.outParam.structField, -1)
				if err != nil {
					log.Fatal(err)
				}
			}
			path.outParam.field = field
		}
	}
	for _, except := range h.exceptMap {
		if except.exceptFunc != nil {
			except.outParam = &outParam{
				httpStatus: http.StatusOK,
			}
			exceptResponse := except.exceptFunc(validErrorCode, "")
			if fn, ok := exceptResponse.(ResponseHeader); ok {
				except.outParam.httpHeader = h.handleHeader(fn.GetHeader())
			}
			if fn, ok := exceptResponse.(ResponseStatusCode); ok {
				except.outParam.httpStatus = fn.GetStatusCode()
			}
			if fn, ok := exceptResponse.(ResponseBody); ok {
				exceptResponse = fn.GetBody()
			}
			fType := reflect.TypeOf(exceptResponse)
			except.outParam.structField = reflect.StructField{Type: fType}
			field, err = h.handleField(except.outParam.structField, -1)
			if err != nil {
				log.Fatal(err)
			}
			except.outParam.field = field
		}
	}
	err = h.handleStruct()
	if err != nil {
		log.Fatal(err)
	}
	// handle other
	for _, path := range h.paths {
		if path.inFs != nil {
			continue
		}
		for _, in := range path.inParams {
			if in.inType == inTypeBody {
				isBody := false
				for _, val := range in.values {
					if !val.mediaType.IsStream() {
						isBody = true
					}
				}
				if isBody {

				}
				continue
			}
		}
		if path.outParam != nil {
			if _, ok := getTypeByCovertInterface[io.ReadCloser](path.outParam.structField.Type); !ok &&
				path.outParam.structField.Type != nil && !path.outParam.field.isTextType {
				val := reflect.New(path.outParam.structField.Type).Elem()
				isNoSupport := h.setExample(val, path.outParam.field, false)
				if isNoSupport {
					path.outParam.example = val.Interface()
				}
			}
		}
	}
	for _, except := range h.exceptMap {
		if except.exceptFunc != nil {
			if _, ok := getTypeByCovertInterface[io.ReadCloser](except.outParam.structField.Type); !ok &&
				except.outParam.structField.Type != nil && !except.outParam.field.isTextType {
				val := reflect.New(except.outParam.structField.Type).Elem()
				isNoSupport := h.setExample(val, except.outParam.field, false)
				if isNoSupport {
					except.outParam.example = val.Interface()
				}
			}
		}
	}
	h.handleOpenapiName()
}

func (h *handler) setExample(val reflect.Value, field *paramField, onlyFind bool, useStructMaps ...map[string]struct{}) (isNoSupport bool) {
	name := field.names.getFieldName(XML)
	for _, v := range name.split {
		switch v {
		case "attr":
		default:
			isNoSupport = true
		}
	}
	useStructMap := map[string]struct{}{}
	if len(useStructMaps) > 0 {
		useStructMap = useStructMaps[0]
	}
	realVal := val
	for val.Kind() == reflect.Ptr {
		initPtr(val)
		val = val.Elem()
	}
	var example any
	if field.tag.example != nil {
		example = field.tag.example
	} else if field.tag._default != nil {
		example = field.tag._default
	} else if len(field.tag.enum) > 0 {
		example = field.tag.enum[0]
	}
	if example != nil && !onlyFind {
		exampleVal := reflect.ValueOf(example)
		for exampleVal.Kind() == reflect.Ptr {
			if exampleVal.IsNil() {
				continue
			}
			exampleVal = exampleVal.Elem()
		}
		if !isNormalType(exampleVal.Type()) {
			isNoSupport = true
		}
		if exampleVal.Type().ConvertibleTo(val.Type()) {
			val.Set(exampleVal.Convert(val.Type()))
			onlyFind = true
		}
	}
	if field.isTextType {
		return
	}
	switch val.Kind() {
	case reflect.Struct:
		if field.pkgName != "" {
			if _, ok := useStructMap[field.pkgName]; ok {
				realVal.Set(reflect.Zero(realVal.Type()))
				return
			}
			useStructMap[field.pkgName] = struct{}{}
		}
		if onlyFind {
			isNoSupport = true
		}
		fields := field.fields
		if field.pkgName != "" {
			stInfo := h.structs[field.pkgName]
			fields = stInfo.fields
		}
		for _, cField := range fields {
			isChildNoSupport := h.setExample(val.Field(cField.index), cField, onlyFind, useStructMap)
			if isChildNoSupport {
				isNoSupport = true
			}
		}
	case reflect.Slice, reflect.Array:
		newVal := reflect.MakeSlice(val.Type(), 1, 1)
		for i := 0; i < newVal.Len(); i++ {
			isChildNoSupport := h.setExample(newVal.Index(i), field.fields[0], onlyFind, useStructMap)
			if isChildNoSupport {
				isNoSupport = true
			}
		}
		if !onlyFind {
			val.Set(newVal)
		}
	case reflect.Map:
		isNoSupport = true
		mapVal := reflect.New(val.Type().Elem()).Elem()
		isChildNoSupport := h.setExample(mapVal, field.fields[0], onlyFind, useStructMap)
		if isChildNoSupport {
			isNoSupport = true
		}
		if !onlyFind {
			mapKey := reflect.New(val.Type().Key()).Elem()
			h.setExample(mapKey, &paramField{tag: &paramTag{}}, false, useStructMap)
			newVal := reflect.MakeMap(reflect.MapOf(mapKey.Type(), mapVal.Type()))
			newVal.SetMapIndex(mapKey, mapVal)
			val.Set(newVal)
		}
	case reflect.String:
		if !onlyFind {
			if field.tag.regexp != "" {
				val.SetString(field.tag.regexp)
			} else {
				val.SetString("string")
			}
		}
	default:
		if !isNumberType(val.Type()) {
			return
		}
		var valFloat *float64
		if field.tag.gt != nil && field.tag.gte != nil {
			if *field.tag.gt > *field.tag.gte {
				valFloat = toPtr(*field.tag.gt + 1)
			} else {
				valFloat = toPtr(*field.tag.gte)
			}
			return
		} else if field.tag.gt != nil {
			valFloat = toPtr(*field.tag.gt + 1)
		} else if field.tag.gte != nil {
			valFloat = toPtr(*field.tag.gte)
		}
		if valFloat == nil {
			if field.tag.lt != nil && field.tag.lte != nil {
				if *field.tag.lt > *field.tag.lte {
					valFloat = toPtr(*field.tag.lte)
				} else {
					valFloat = toPtr(*field.tag.lt - 1)
				}
			} else if field.tag.lt != nil {
				valFloat = toPtr(*field.tag.lt - 1)
			} else if field.tag.lte != nil {
				valFloat = toPtr(*field.tag.lte)
			}
		}
		if valFloat == nil {
			return
		}
		val.Set(reflect.ValueOf(*valFloat).Convert(val.Type()))
	}
	return
}

func (h *handler) handleOutParam(outParam *outParam) {
	fType := outParam.structField.Type
	var value reflect.Value
	if fType.Kind() == reflect.Ptr {
		value = reflect.New(fType.Elem())
	} else {
		value = reflect.New(fType).Elem()
	}
	valAny := value.Interface()
	if fn, ok := valAny.(ResponseHeader); ok {
		outParam.httpHeader = h.handleHeader(fn.GetHeader())
	}
	if fn, ok := valAny.(ResponseStatusCode); ok {
		outParam.httpStatus = fn.GetStatusCode()
	}
	if fn, ok := valAny.(ResponseBody); ok {
		outParam.structField.Type = reflect.TypeOf(fn.GetBody())
	}
}

func (h *handler) handleHeader(header http.Header) http.Header {
	rs := http.Header{}
	for key, val := range header {
		key = textproto.CanonicalMIMEHeaderKey(key)
		rs[key] = val
	}
	contentType := MediaType(rs.Get("Content-Type"))
	if contentType.Tag() != "" {
		h.mediaTypes[contentType.MediaType()] = struct{}{}
	}
	return rs
}

func (h *handler) handleParam(inType InType, field reflect.StructField, index int, names paramFieldNames) (rs *paramField, err error) {
	rs = &paramField{
		index: index,
		names: names,
		tag:   &paramTag{},
	}
	fType := field.Type
	fType = removeMorePtr(fType)
	eType := fType
	if fType.Kind() == reflect.Ptr {
		eType = fType.Elem()
	}
	rs._type = eType
	rs.kind = eType.Kind()
	switch inType {
	case inTypeFile:
		if fType.ConvertibleTo(typeFile) {
			rs._type = fType
			rs.kind = reflect.String
		}
	case inTypeCookie:
		if fType.ConvertibleTo(typeCookie) {
			rs._type = fType
			rs.kind = reflect.String
		}
	}
	if isTextInterface(fType) {
		rs._type = fType
		rs.kind = reflect.String
		rs.isTextType = true
	}
	fVal := getValueByType(fType, true)
	if err = h.handleTagByInterface(fType, rs, fVal); err != nil {
		return
	}
	if err = h.handleTagByField(field, rs); err != nil {
		return
	}
	h.handleTagByType(eType.Kind(), rs.tag)
	if rs.isTextType {
		return
	}
	var childField *paramField
	switch eType.Kind() {
	case reflect.Slice, reflect.Array:
		childField, err = h.handleParam(inType, reflect.StructField{Type: eType.Elem()}, -1, names)
		if err != nil {
			return
		}
		rs.fields = append(rs.fields, childField)
	default:
	}
	return
}

func (h *handler) handleField(field reflect.StructField, index int, beforeStructPkgName ...string) (rs *paramField, err error) {
	names := h.getNames(field)
	if len(names) == 0 {
		return
	}
	rs = &paramField{
		index:     index,
		tag:       &paramTag{},
		name:      field.Name,
		names:     names,
		anonymous: field.Anonymous,
	}
	fType := field.Type
	fType = removeMorePtr(fType)
	eType := fType
	if fType.Kind() == reflect.Ptr {
		eType = fType.Elem()
	}
	if rs.name == "" {
		rs.name = eType.Name()
	}
	rs._type = eType
	rs.kind = eType.Kind()
	rs.pkgName = h.getPkgName(eType)
	if isTextInterface(fType) {
		rs._type = fType
		rs.kind = reflect.String
		rs.isTextType = true
	}
	fVal := getValueByType(fType, true)
	if err = h.handleTagByInterface(fType, rs, fVal); err != nil {
		return
	}
	if err = h.handleTagByField(field, rs); err != nil {
		return
	}
	h.handleTagByType(eType.Kind(), rs.tag)
	if rs.isTextType {
		return
	}
	var childField *paramField
	switch eType.Kind() {
	case reflect.Map:
		var childKey *paramField
		childKey, err = h.handleField(reflect.StructField{Type: eType.Key()}, -1)
		if err != nil {
			return
		}
		if childKey.kind != reflect.String {
			err = fmt.Errorf("the key of a map must be of string type")
			return
		}
		childField, err = h.handleField(reflect.StructField{Type: eType.Elem()}, -1, beforeStructPkgName...)
		if err != nil {
			return
		}
		rs.fields = append(rs.fields, childKey, childField)
	case reflect.Slice, reflect.Array:
		childField, err = h.handleField(reflect.StructField{Type: eType.Elem()}, -1, beforeStructPkgName...)
		if err != nil {
			return
		}
		rs.fields = append(rs.fields, childField)
	case reflect.Struct:
		if rs.pkgName == "" {
			var pField *paramField
			for i := 0; i < eType.NumField(); i++ {
				vField := eType.Field(i)
				if vField.Name[0] < 'A' || vField.Name[0] > 'Z' {
					continue
				}
				pField, err = h.handleField(vField, i, beforeStructPkgName...)
				if err != nil {
					return
				}
				if pField == nil {
					continue
				}
				if vField.Name == "XMLName" {
					nameSplit := strings.Split(vField.Tag.Get("xml"), ",")
					if nameSplit[0] != "" {
						rs.xmlName = nameSplit[0]
					}
				}
				rs.fields = append(rs.fields, pField)
			}
			return
		}
		if len(beforeStructPkgName) > 0 && !inArray(rs.pkgName, h.structDepends[beforeStructPkgName[0]]) {
			h.structDepends[beforeStructPkgName[0]] = append(h.structDepends[beforeStructPkgName[0]], rs.pkgName)
		}

		if _, ok := h.structs[rs.pkgName]; !ok {
			h.structTypes[rs.pkgName] = fType
		}
	default:
	}
	return
}

func (h *handler) getNames(field reflect.StructField) (rs paramFieldNames) {
	for mediaType := range h.mediaTypes {
		name := field.Tag.Get(mediaType.Tag())
		if name == "-" {
			continue
		}
		nameSplit := strings.Split(name, ",")
		name = nameSplit[0]
		if name == "" {
			name = mediaType.DefaultName(field.Name)
		}
		paramName := paramFieldName{
			name:      name,
			mediaType: mediaType,
			required:  true,
		}
		for _, v := range nameSplit[1:] {
			if v == omitempty {
				paramName.required = false
				continue
			}
			paramName.split = append(paramName.split, v)
		}
		rs = append(rs, paramName)
	}
	return
}

func (h *handler) handleStruct() (err error) {
	var pField *paramField
	for len(h.structTypes) > 0 {
		for pkgName, structType := range h.structTypes {
			stInfo := &structInfo{
				_type:   structType,
				xmlName: structType.Name(),
			}
			for i := 0; i < structType.NumField(); i++ {
				field := structType.Field(i)
				if field.Name[0] < 'A' || field.Name[0] > 'Z' {
					continue
				}
				pField, err = h.handleField(field, i, pkgName)
				if err != nil {
					return
				}
				if pField == nil {
					continue
				}
				if field.Name == "XMLName" {
					nameSplit := strings.Split(field.Tag.Get("xml"), ",")
					if nameSplit[0] != "" {
						stInfo.xmlName = nameSplit[0]
					}
				}
				stInfo.fields = append(stInfo.fields, pField)
			}
			h.structs[pkgName] = stInfo
			delete(h.structTypes, pkgName)
		}
	}
	return
}

func (h *handler) handleTagByInterface(fType reflect.Type, field *paramField, valPtr reflect.Value) (err error) {
	var val any
	if fType.Kind() == reflect.Ptr {
		val = valPtr.Interface()
		fType = fType.Elem()
	} else {
		val = valPtr.Elem().Interface()
	}
	kind := field.kind
	if iTag, ok := val.(TagRegexp); ok && kind == reflect.String {
		field.tag.regexp = iTag.Regexp()
	}
	if iTag, ok := val.(TagEnum); ok && isNormalType(fType) {
		field.tag.enum = iTag.Enum()
		if err = h.handleTagEnumToFloat64(field.tag.enum, fType); err != nil {
			return
		}
	}
	if iTag, ok := val.(TagLt); ok && isNumberType(fType) {
		field.tag.lt = toPtr(iTag.Lt())
	}
	if iTag, ok := val.(TagLte); ok && isNumberType(fType) {
		field.tag.lte = toPtr(iTag.Lte())
	}
	if iTag, ok := val.(TagGt); ok && isNumberType(fType) {
		field.tag.gt = toPtr(iTag.Gt())
	}
	if iTag, ok := val.(TagGte); ok && isNumberType(fType) {
		field.tag.gte = toPtr(iTag.Gte())
	}
	if iTag, ok := val.(TagMultiple); ok && isNumberType(fType) {
		field.tag.multiple = toPtr(iTag.Multiple())
	}
	if iTag, ok := val.(TagMax); ok && inArray(kind, []reflect.Kind{reflect.Array, reflect.Slice,
		reflect.Map, reflect.String}) {
		field.tag.max = toPtr(iTag.Max())
	}
	if iTag, ok := val.(TagMin); ok && inArray(kind, []reflect.Kind{reflect.Array, reflect.Slice,
		reflect.Map, reflect.String}) {
		field.tag.min = iTag.Min()
	}
	if iTag, ok := val.(TagUnique); ok && inArray(kind, []reflect.Kind{reflect.Array, reflect.Slice}) {
		field.tag.unique = iTag.Unique()
	}
	if iTag, ok := val.(TagDesc); ok {
		field.tag.desc = h.getMappingTag(iTag.Desc())
	}
	if iTag, ok := val.(TagDefault); ok {
		field.tag._default = iTag.Default()
	}
	if iTag, ok := val.(TagExample); ok {
		field.tag.example = iTag.Example()
	}
	if iTag, ok := val.(TagDeprecated); ok {
		field.tag.deprecated = iTag.Deprecated()
	}
	return
}

func (h *handler) handleTagByField(field reflect.StructField, pField *paramField) (err error) {
	fType := field.Type
	for fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}
	kind := pField.kind
	if tagVal := field.Tag.Get(tagRegexp); tagVal != "" && kind == reflect.String {
		pField.tag.regexp = tagVal
	}
	if tagVal := field.Tag.Get(tagEnum); tagVal != "" && (isNormalType(fType) || kind == reflect.String) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.enum, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagLt); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.lt, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagLte); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.lte, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagGt); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.gt, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagGte); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.gte, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagMultiple); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.multiple, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagMax); tagVal != "" && inArray(kind, []reflect.Kind{reflect.Array,
		reflect.Slice, reflect.Map, reflect.String}) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.max, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagMin); tagVal != "" && inArray(kind, []reflect.Kind{reflect.Array,
		reflect.Slice, reflect.Map, reflect.String}) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.min, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagUnique); tagVal != "" && inArray(kind, []reflect.Kind{reflect.Array,
		reflect.Slice}) {
		if err = h.parseTagValByKind(tagVal, &pField.tag.unique, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagDesc); tagVal != "" {
		pField.tag.desc = h.getMappingTag(tagVal)
	}
	if tagVal := field.Tag.Get(tagDefault); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &pField.tag._default, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagExample); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &pField.tag.example, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagDeprecated); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &pField.tag.deprecated, kind); err != nil {
			return
		}
	}
	return
}

func (h *handler) handleTagByType(kind reflect.Kind, tag *paramTag) {
	minNum, maxNum := float64(0), float64(-1)
	switch kind {
	case reflect.Int8:
		minNum = math.MinInt8
		maxNum = math.MaxInt8
	case reflect.Int16:
		minNum = math.MinInt16
		maxNum = math.MaxInt16
	case reflect.Int32:
		minNum = math.MinInt32
		maxNum = math.MaxInt32
	case reflect.Int64:
		minNum = math.MinInt64
		maxNum = math.MaxInt64
	case reflect.Uint8:
		maxNum = math.MaxUint8
	case reflect.Uint16:
		maxNum = math.MaxUint16
	case reflect.Uint32:
		maxNum = math.MaxUint32
	case reflect.Uint64:
		maxNum = math.MaxUint64
	default:
		return
	}
	if maxNum != -1 {
		if tag.lte == nil || *tag.lte > maxNum {
			tag.lte = toPtr(maxNum)
		}
		if tag.gte == nil || *tag.gte < minNum {
			tag.gte = toPtr(minNum)
		}
	}
}

func (h *handler) getPkgName(fType reflect.Type) string {
	if fType.PkgPath() == "" || fType.Name() == "" {
		return ""
	}
	return fmt.Sprintf("%s.%s", fType.PkgPath(), fType.Name())
}

func (h *handler) handleTagEnumToFloat64(enum []any, fType reflect.Type) (err error) {
	if !isNumberType(fType) {
		return
	}
	for k, v := range enum {
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v = float64(val.Int())
		case reflect.Uint, reflect.Uint8,
			reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v = float64(val.Uint())
		case reflect.Float32, reflect.Float64:
			v = val.Float()
		case reflect.String:
			if v, err = strconv.ParseFloat(val.String(), 64); err != nil {
				return
			}
		case reflect.Bool:
			v = 0
			if val.Bool() {
				v = 1
			}
		default:
			return fmt.Errorf("invalid enum value type: %s", val.Kind())
		}
		enum[k] = v
	}
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
				*val = v
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

func (h *handler) handleOpenapiName() {
	sMap := map[string]int{}
	fMap := map[string]struct{}{}
	for pkgName := range h.structs {
		pkgNameList := h.splitPkgName(pkgName)
		for _, name := range pkgNameList {
			if _, ok := fMap[name]; ok {
				continue
			}
			fMap[name] = struct{}{}
			long, short := h.getShortPkgName(name)
			sMap[long]++
			sMap[short]++
		}
	}
	for pkgName, stInfo := range h.structs {
		pkgNameList := h.splitPkgName(pkgName)
		for _, name := range pkgNameList {
			long, short := h.getShortPkgName(name)
			if sMap[short] == 1 {
				pkgName = strings.ReplaceAll(pkgName, name, short)
				continue
			}
			if sMap[long] == 1 {
				pkgName = strings.ReplaceAll(pkgName, name, long)
				continue
			}
		}
		stInfo.openapiName = strings.NewReplacer(
			"/", ".",
			"[", "_",
			"]", "",
		).Replace(pkgName)
	}
}

func (h *handler) splitPkgName(pkgName string) (rs []string) {
	for pkgName[len(pkgName)-1] == ']' {
		leftNum := 0
		symbolIndex := 0
		for i := len(pkgName) - 2; i >= 0; i-- {
			if pkgName[i] == ']' {
				leftNum++
			} else if pkgName[i] == '[' {
				if leftNum == 0 {
					symbolIndex = i
					break
				}
				leftNum--
			}
		}
		rs = append(rs, pkgName[:symbolIndex])
		pkgName = pkgName[symbolIndex+1 : len(pkgName)-1]
	}
	rs = append(rs, pkgName)
	return
}

func (h *handler) getShortPkgName(pkgName string) (long, short string) {
	prefix := ""
	if pkgName[0] == '[' {
		symbolIndex := strings.IndexByte(pkgName, ']')
		prefix = pkgName[:symbolIndex+1]
		pkgName = pkgName[symbolIndex+1:]
	}
	pkgNameList := strings.Split(pkgName, "/")
	long = prefix + pkgNameList[len(pkgNameList)-1]
	pkgNameList = strings.Split(long, ".")
	short = prefix + pkgNameList[len(pkgNameList)-1]
	return
}

func (h *handler) getMappingTag(tagVal string, replaces ...map[string]struct{}) string {
	if tagVal == "" {
		return tagVal
	}
	replace := map[string]struct{}{}
	if len(replaces) > 0 {
		replace = replaces[0]
	}
	tagList := h.handleMappingTag(tagVal)
	newTagVal := ""
	for _, tag := range tagList {
		c := tag[0]
		tag = tag[1:]
		if c == '1' {
			oldVal := tag[2 : len(tag)-2]
			val := h.api.structTagVariableMap[oldVal]
			if val != nil {
				if _, ok := replace[oldVal]; ok {
					log.Fatal(fmt.Sprintf("mapping tag '%v' dead loop", oldVal))
				}
				newTagVal += h.getMappingTag(val.(string), h.cloneMappingAppend(replace, oldVal))
				continue
			}
		}
		newTagVal += tag
	}
	return newTagVal
}

func (h *handler) cloneMappingAppend(m map[string]struct{}, key string) map[string]struct{} {
	rs := map[string]struct{}{}
	for k, v := range m {
		rs[k] = v
	}
	rs[key] = struct{}{}
	return rs
}

func (h *handler) handleMappingTag(tagVal string) []string {
	n := len(tagVal)
	if n < 5 {
		return []string{"0" + tagVal}
	}
	var list []string
	left, right := -1, 0
	for i := 0; i < n; i++ {
		if left == -1 {
			if i+2 < n && tagVal[i] == '{' && tagVal[i+1] == '{' && tagVal[i+2] != '{' {
				left = i
			}
			continue
		}
		if i-2 >= 0 && tagVal[i-2] != '}' && tagVal[i-1] == '}' && tagVal[i] == '}' {
			if right < left {
				if right == -1 {
					list = append(list, "0"+tagVal[:left])
				} else {
					list = append(list, "0"+tagVal[right:left])
				}
			}
			right = i + 1
			list = append(list, "1"+tagVal[left:right])
			left = -1
			continue
		}
	}
	if right < n {
		list = append(list, "0"+tagVal[right:])
	}
	return list
}
