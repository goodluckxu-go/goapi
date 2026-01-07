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

	"github.com/goodluckxu-go/goapi/openapi"
	"github.com/goodluckxu-go/goapi/swagger"
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
	except                 *outParam
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
	for k, v := range obj.docsMap {
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
			if in.inType == inTypeFile {
				if !isArrayType(in.structField.Type, func(sType reflect.Type) bool {
					if sType.ConvertibleTo(typeFile) {
						return true
					}
					return false
				}, 2) {
					log.Fatal(fmt.Sprintf("the type of parameter '%v' in '%v' must be "+
						"‘*multipart.FileHeade’ or an array of ‘*multipart.FileHeader’, has type '%v'",
						in.values[0].name, in.inType.Tag(),
						in.structField.Type.String()))
				}
			} else if in.inType == inTypeBody {
				isBody := false
				for _, val := range in.values {
					if val.mediaType.IsStream() {
						if !isArrayType(in.structField.Type, func(sType reflect.Type) bool {
							if sType.ConvertibleTo(typeBytes) || sType.Kind() == reflect.String {
								return true
							}
							if _, ok := getTypeByCovertInterface[io.ReadCloser](sType); ok {
								return true
							}
							return false
						}, 1) {
							log.Fatal("other media types only support types '[]byte', 'string', and 'io.ReadCloser‘")
						}
					} else {
						isBody = true
					}
				}
				if isBody {
					field, err = h.handleField(in.structField, -1)
					if err != nil {
						log.Fatal(err)
					}
					field.anonymous = true
					in.field = field
				} else {
					in.field = &paramField{
						tag: &paramTag{},
					}
					fType := removeMorePtr(in.structField.Type)
					in.field.kind = fType.Kind()
					if fType.Kind() == reflect.Ptr {
						in.field.kind = fType.Elem().Kind()
					}
					fVal := getValueByType(fType, true)
					if err = h.handleTagByInterface(fType, in.field.tag, fVal); err != nil {
						log.Fatal(err)
					}
					if _, ok := getTypeByCovertInterface[io.ReadCloser](fType); ok {
						in.field._type = fType
					}
					if fType.Kind() == reflect.Ptr {
						fType = fType.Elem()
					}
					if in.field._type == nil {
						in.field._type = fType
					}
					if err = h.handleTagByField(in.structField, in.field.tag, fVal); err != nil {
						log.Fatal(err)
					}
					if _, ok := getTypeByCovertInterface[TextInterface](fVal); ok {
						in.field.kind = reflect.String
					}
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
					if _, ok := getTypeByCovertInterface[TextInterface](sType, true); ok {
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
	if h.api.exceptFunc != nil {
		exceptResponse := h.api.exceptFunc(validErrorCode, "")
		fType := reflect.TypeOf(exceptResponse.GetBody())
		h.except = &outParam{
			structField: reflect.StructField{Type: fType},
			httpStatus:  exceptResponse.GetStatusCode(),
			httpHeader:  h.handleHeader(exceptResponse.GetHeader()),
		}
		field, err = h.handleField(h.except.structField, -1)
		if err != nil {
			log.Fatal(err)
		}
		h.except.field = field
	}
	err = h.handleStruct()
	if err != nil {
		log.Fatal(err)
	}
	h.handleOpenapiName()
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
	fVal := getValueByType(fType, true)
	if err = h.handleTagByInterface(fType, rs.tag, fVal); err != nil {
		return
	}
	if err = h.handleTagByField(field, rs.tag, fVal); err != nil {
		return
	}
	if fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}
	h.handleTagByType(fType.Kind(), rs.tag)
	if rs._type == nil {
		rs._type = fType
	}
	if rs.kind == 0 {
		rs.kind = fType.Kind()
	}
	if rType, ok := getTypeByCovertInterface[TextInterface](fVal); ok {
		rs._type = rType
		rs.kind = reflect.String
		rs.isTextType = true
		return
	}
	var childField *paramField
	switch fType.Kind() {
	case reflect.Slice, reflect.Array:
		childField, err = h.handleParam(inType, reflect.StructField{Type: fType.Elem()}, -1, names)
		if err != nil {
			return
		}
		rs._type = fType
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
		names:     names,
		anonymous: field.Anonymous,
	}
	fType := field.Type
	fType = removeMorePtr(fType)
	fVal := getValueByType(fType, true)
	if err = h.handleTagByInterface(fType, rs.tag, fVal); err != nil {
		return
	}
	if fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}
	if err = h.handleTagByField(field, rs.tag, fVal); err != nil {
		return
	}
	h.handleTagByType(fType.Kind(), rs.tag)
	rs._type = fType
	rs.kind = fType.Kind()
	rs.pkgName = h.getPkgName(fType)
	if rType, ok := getTypeByCovertInterface[TextInterface](fVal); ok {
		rs._type = rType
		rs.kind = reflect.String
		rs.isTextType = true
		return
	}
	var childField *paramField
	switch fType.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		childField, err = h.handleField(reflect.StructField{Type: fType.Elem()}, -1, beforeStructPkgName...)
		if err != nil {
			return
		}
		rs.fields = append(rs.fields, childField)
	case reflect.Struct:
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
				_type: structType,
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
				stInfo.fields = append(stInfo.fields, pField)
			}
			h.structs[pkgName] = stInfo
			delete(h.structTypes, pkgName)
		}
	}
	return
}

func (h *handler) handleTagByInterface(fType reflect.Type, tag *paramTag, valPtr reflect.Value) (err error) {
	var val any
	if fType.Kind() == reflect.Ptr {
		val = valPtr.Interface()
		fType = fType.Elem()
	} else {
		val = valPtr.Elem().Interface()
	}
	kind := fType.Kind()
	isCustomType := false
	if _, ok := getTypeByCovertInterface[TextInterface](valPtr, true); ok {
		kind = reflect.String
		isCustomType = true
	}
	if iTag, ok := val.(TagRegexp); ok && kind == reflect.String {
		tag.regexp = iTag.Regexp()
	}
	if iTag, ok := val.(TagEnum); ok && isNormalType(fType) {
		tag.enum = iTag.Enum()
		if err = h.handleTagEnumToFloat64(tag.enum, fType); err != nil {
			return
		}
	}
	if iTag, ok := val.(TagLt); ok && isNumberType(fType) {
		tag.lt = toPtr(iTag.Lt())
	}
	if iTag, ok := val.(TagLte); ok && isNumberType(fType) {
		tag.lte = toPtr(iTag.Lte())
	}
	if iTag, ok := val.(TagGt); ok && isNumberType(fType) {
		tag.gt = toPtr(iTag.Gt())
	}
	if iTag, ok := val.(TagGte); ok && isNumberType(fType) {
		tag.gte = toPtr(iTag.Gte())
	}
	if iTag, ok := val.(TagMultiple); ok && isNumberType(fType) {
		tag.multiple = toPtr(iTag.Multiple())
	}
	if iTag, ok := val.(TagMax); ok && inArray(kind, []reflect.Kind{reflect.Array, reflect.Slice,
		reflect.Map, reflect.String}) {
		tag.max = toPtr(iTag.Max())
	}
	if iTag, ok := val.(TagMin); ok && inArray(kind, []reflect.Kind{reflect.Array, reflect.Slice,
		reflect.Map, reflect.String}) {
		tag.min = iTag.Min()
	}
	if iTag, ok := val.(TagUnique); ok && inArray(kind, []reflect.Kind{reflect.Array, reflect.Slice}) {
		tag.unique = iTag.Unique()
	}
	if iTag, ok := val.(TagDesc); ok {
		tag.desc = h.getMappingTag(iTag.Desc())
	}
	if iTag, ok := val.(TagDefault); ok {
		valAny := iTag.Default()
		if isCustomType {
			fVal := reflect.ValueOf(valAny).Convert(fType)
			if fn, fnOk := getFnByCovertInterface[TextInterface](fVal, true); fnOk {
				var valBytes []byte
				if valBytes, err = fn.MarshalText(); err != nil {
					return
				}
				valAny = string(valBytes)
			}
		}
		tag._default = valAny
	}
	if iTag, ok := val.(TagExample); ok {
		valAny := iTag.Example()
		if isCustomType {
			fVal := reflect.ValueOf(valAny).Convert(fType)
			if fn, fnOk := getFnByCovertInterface[TextInterface](fVal, true); fnOk {
				var valBytes []byte
				if valBytes, err = fn.MarshalText(); err != nil {
					return
				}
				valAny = string(valBytes)
			}
		}
		tag.example = valAny
	}
	if iTag, ok := val.(TagDeprecated); ok {
		tag.deprecated = iTag.Deprecated()
	}
	return
}

func (h *handler) handleTagByField(field reflect.StructField, tag *paramTag, valPtr reflect.Value) (err error) {
	fType := field.Type
	for fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}
	kind := fType.Kind()
	if _, ok := getTypeByCovertInterface[TextInterface](valPtr, true); ok {
		kind = reflect.String
	}
	if tagVal := field.Tag.Get(tagRegexp); tagVal != "" && kind == reflect.String {
		tag.regexp = tagVal
	}
	if tagVal := field.Tag.Get(tagEnum); tagVal != "" && (isNormalType(fType) || kind == reflect.String) {
		if err = h.parseTagValByKind(tagVal, &tag.enum, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagLt); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &tag.lt, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagLte); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &tag.lte, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagGt); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &tag.gt, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagGte); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &tag.gte, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagMultiple); tagVal != "" && isNumberType(fType) {
		if err = h.parseTagValByKind(tagVal, &tag.multiple, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagMax); tagVal != "" && inArray(kind, []reflect.Kind{reflect.Array,
		reflect.Slice, reflect.Map, reflect.String}) {
		if err = h.parseTagValByKind(tagVal, &tag.max, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagMin); tagVal != "" && inArray(kind, []reflect.Kind{reflect.Array,
		reflect.Slice, reflect.Map, reflect.String}) {
		if err = h.parseTagValByKind(tagVal, &tag.min, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagUnique); tagVal != "" && inArray(kind, []reflect.Kind{reflect.Array,
		reflect.Slice}) {
		if err = h.parseTagValByKind(tagVal, &tag.unique, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagDesc); tagVal != "" {
		tag.desc = h.getMappingTag(tagVal)
	}
	if tagVal := field.Tag.Get(tagDefault); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &tag._default, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagExample); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &tag.example, kind); err != nil {
			return
		}
	}
	if tagVal := field.Tag.Get(tagDeprecated); tagVal != "" {
		if err = h.parseTagValByKind(tagVal, &tag.deprecated, kind); err != nil {
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
