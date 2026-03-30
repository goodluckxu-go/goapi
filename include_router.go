package goapi

import (
	"fmt"
	"net/textproto"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/goodluckxu-go/goapi/v2/openapi"
)

type includeRouter struct {
	router      any
	prefix      string
	groupPrefix string
	isDocs      bool
	docsPath    string
	childPath   string
	middlewares []HandleFunc
}

func (i *includeRouter) returnObj() (obj returnObjResult, err error) {
	obj.docsMap = map[string]returnObjDocs{
		i.docsPath: {},
	}
	obj.mediaTypes = map[MediaType]struct{}{}
	var tags []*openapi.Tag
	var tagStrs []string
	if tVal, ok := i.router.(RouterTags); ok {
		tags = tVal.Tags()
		for _, tag := range tags {
			tagStrs = append(tagStrs, tag.Name)
		}
	}
	if i.isDocs && len(tags) > 0 {
		obj.docsMap = map[string]returnObjDocs{
			i.docsPath: {
				tags: tags,
			},
		}
	}
	value := reflect.ValueOf(i.router)
	var pInfo *pathInfo
	if value.Kind() == reflect.Func {
		funcPos := runtime.FuncForPC(value.Pointer()).Name()
		pInfo, err = i.handleRouter(value)
		if err != nil {
			err = fmt.Errorf("%v, pos: %v", err, funcPos)
			return
		}
		pInfo.pos = funcPos
		for _, v := range pInfo.inParams {
			for _, mediaType := range v.values.MediaTypes() {
				if mediaType.Tag() != "" {
					obj.mediaTypes[mediaType] = struct{}{}
				}
			}
		}
		pInfo.tags = tagStrs
		obj.paths = append(obj.paths, pInfo)
		return
	}
	pos := fmt.Sprintf("%v.%v", value.Type().PkgPath(), value.Type().Name())
	if value.Kind() == reflect.Ptr {
		pos = fmt.Sprintf("%v.(*%v)", value.Elem().Type().PkgPath(), value.Elem().Type().Name())
	}
	if !((value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Struct) || value.Kind() == reflect.Struct) {
		err = fmt.Errorf("router must be a struct, struct pointer or function, pos: %v", pos)
		return
	}
	numMethod := value.NumMethod()
	for j := 0; j < numMethod; j++ {
		funcPos := fmt.Sprintf("%v.%v", pos, value.Type().Method(j).Name)
		routerMethod := value.Method(j)
		pInfo, err = i.handleRouter(routerMethod)
		if err != nil {
			err = fmt.Errorf("%v, pos: %v", err, funcPos)
			return
		}
		if pInfo == nil {
			continue
		}
		pInfo.pos = funcPos
		for _, v := range pInfo.inParams {
			for _, mediaType := range v.values.MediaTypes() {
				if mediaType.Tag() != "" {
					obj.mediaTypes[mediaType] = struct{}{}
				}
			}
		}
		pInfo.tags = tagStrs
		obj.paths = append(obj.paths, pInfo)
	}
	return
}

func (i *includeRouter) handleRouter(routerMethod reflect.Value) (pInfo *pathInfo, err error) {
	// handle in param
	numIn := routerMethod.Type().NumIn()
	if numIn == 0 {
		return
	}
	inTypes := make([]reflect.Type, numIn)
	var inputType reflect.Type
	switch numIn {
	case 1:
		if routerMethod.Type().In(0).Kind() != reflect.Struct {
			err = fmt.Errorf("when the method parameter in the router must be 1, it must be a structure")
			return
		}
		inputType = routerMethod.Type().In(0)
		inTypes[0] = inputType
	case 2:
		cType := reflect.TypeOf(&Context{})
		if routerMethod.Type().In(0) != cType {
			err = fmt.Errorf("when the method parameter in the router must be 2, the 1st parameter must be '*goapi.Context'")
			return
		}
		inTypes[0] = cType
		if routerMethod.Type().In(1).Kind() != reflect.Struct {
			err = fmt.Errorf("when the method parameter in the router must be 2, the 2st parameter must be a structure")
			return
		}
		inputType = routerMethod.Type().In(1)
		inTypes[1] = inputType
	default:
		err = fmt.Errorf("the method parameters in the router must be 1 or 2")
		return
	}
	pInfo = &pathInfo{
		value:       routerMethod,
		inTypes:     inTypes,
		middlewares: i.middlewares,
		isDocs:      i.isDocs,
		docsPath:    i.docsPath,
		childPath:   i.childPath,
		groupPrefix: i.groupPrefix,
	}
	numField := inputType.NumField()
	var in []*inParam
	allMds := allMethods()
	for l := 0; l < numField; l++ {
		field := inputType.Field(l)
		switch field.Type {
		case reflect.TypeOf(Router{}):
			pathStr, pathOk := field.Tag.Lookup(tagPaths)
			methodStr, methodOk := field.Tag.Lookup(tagMethods)
			if (i.prefix == "" && pathStr == "") || methodStr == "" || !pathOk || !methodOk {
				err = fmt.Errorf("the 'goapi.Router' parameter must have tags for 'paths' and 'methods'")
				return
			}
			methods := strings.Split(methodStr, ",")
			upperMethods := make([]string, len(methods))
			for k, v := range methods {
				upperMethods[k] = strings.ToUpper(v)
			}
			var noMethods []string
			for k, v := range upperMethods {
				if !inArray(v, allMds) {
					noMethods = append(noMethods, methods[k])
				}
			}
			if len(noMethods) > 0 {
				err = fmt.Errorf("methods '%v' does not exist, must be in '%v'",
					strings.Join(noMethods, "', '"), strings.Join(allMds, "', '"))
				return
			}
			paths := strings.Split(pathStr, ",")
			for k, v := range paths {
				paths[k] = pathJoin(i.prefix, v)
			}
			deprecated := false
			deprecatedStr := field.Tag.Get("deprecated")
			if deprecatedStr != "" {
				if deprecated, err = strconv.ParseBool(field.Tag.Get("deprecated")); err != nil {
					return
				}
			}
			pInfo.paths = paths
			pInfo.methods = upperMethods
			pInfo.summary = field.Tag.Get(tagSummary)
			pInfo.desc = field.Tag.Get(tagDesc)
			pInfo.deprecated = deprecated
			tag := field.Tag.Get(tagTags)
			if tag != "" {
				pInfo.tags = strings.Split(tag, ",")
			}
		default:
			if in, err = i.parseIn(field, []int{l}, ""); err != nil {
				return
			}
			pInfo.inParams = append(pInfo.inParams, in...)
		}
	}
	if len(pInfo.paths) == 0 || len(pInfo.methods) == 0 {
		err = fmt.Errorf("the 'goapi.Router' parameter must exist")
		return
	}
	// handle out param
	numOut := routerMethod.Type().NumOut()
	if numOut > 1 {
		err = fmt.Errorf("returns at most one result, and there are %v results", numOut)
		return
	}
	if numOut == 1 {
		outType := routerMethod.Type().Out(0)
		pInfo.outParam = &outParam{
			structField: reflect.StructField{Type: outType},
		}
	}
	return
}

func (i *includeRouter) parseIn(field reflect.StructField, deeps []int, securityType InType) (params []*inParam, err error) {
	if field.Name[0] < 'A' || field.Name[0] > 'Z' {
		return
	}
	// handle security
	var securityInType InType
	if field.Type.Implements(securityTypeHTTPBasic) {
		securityInType = inTypeSecurityHTTPBasic
	}
	if field.Type.Implements(securityTypeHTTPBearer) {
		if securityInType != "" {
			err = fmt.Errorf("field %v cannot implement both '%v' and '%v' simultaneously", field.Name,
				securityInType, inTypeSecurityHTTPBearer)
			return
		}
		securityInType = inTypeSecurityHTTPBearer
	}
	if field.Type.Implements(securityTypeHTTPBearerJWT) {
		if securityInType != "" {
			err = fmt.Errorf("field %v cannot implement both '%v' and '%v' simultaneously", field.Name,
				securityInType, inTypeSecurityHTTPBearerJWT)
			return
		}
		securityInType = inTypeSecurityHTTPBearerJWT
	}
	if field.Type.Implements(securityTypeApiKey) {
		if securityInType != "" {
			err = fmt.Errorf("field %v cannot implement both '%v' and '%v' simultaneously", field.Name,
				securityInType, inTypeSecurityApiKey)
			return
		}
		securityInType = inTypeSecurityApiKey
	}
	if securityInType != "" {
		params = append(params, &inParam{
			inType:      securityInType,
			structField: field,
			deeps:       deeps,
		})
		if securityType != "" {
			err = fmt.Errorf("security ‘%v’ cannot exist under security '%v'", securityInType, securityType)
			return
		}
	}
	// handle param
	in := &inParam{structField: field, deeps: deeps, parentInType: securityType}
	for _, inType := range InType("").List() {
		inTypeValue := field.Tag.Get(inType.Tag())
		if inTypeValue != "" {
			if in.inType != "" {
				err = fmt.Errorf("field %s cannot have both '%s' and '%s' labels present at the same time",
					field.Name, in.inType.Tag(), inType.Tag())
				return
			}
			in.inType = inType
			valSplit := strings.Split(inTypeValue, ",")
			if in.inType == inTypeBody {
				for _, val := range valSplit {
					valType := MediaType(val)
					if valType.Tag() == JSON.Tag() {
						valType = JSON
					} else if valType.Tag() == XML.Tag() {
						valType = XML
					} else if valType.Tag() == YAML.Tag() {
						valType = YAML
					}
					if !inArray(valType, in.values.MediaTypes()) {
						in.values = append(in.values, paramFieldName{mediaType: valType})
					}
				}
			} else {
				name := valSplit[0]
				if in.inType == inTypeHeader {
					name = textproto.CanonicalMIMEHeaderKey(name)
				}
				value := paramFieldName{required: true, name: name, inType: in.inType}
				for _, val := range valSplit[1:] {
					if val == omitempty {
						value.required = false
					} else {
						value.split = append(value.split, val)
					}
				}
				in.values = paramFieldNames{value}
			}
			params = append(params, in)
		}
	}
	if in.inType != "" {
		return
	}
	// handle struct recursion
	fType := field.Type
	for fType.Kind() == reflect.Ptr {
		if fType == typeContext {
			in.inType = inTypeOther
			params = append(params, in)
			return
		}
		fType = fType.Elem()
	}
	if fType.Kind() != reflect.Struct {
		return
	}
	numField := fType.NumField()
	var childParams []*inParam
	for j := 0; j < numField; j++ {
		if childParams, err = i.parseIn(fType.Field(j), append(deeps, j), securityInType); err != nil {
			return
		}
		for _, item := range childParams {
			item.parentName = field.Name
		}
		params = append(childParams, params...)
	}
	return
}
