package goapi

import (
	"fmt"
	"net/textproto"
	"reflect"
	"strings"

	"github.com/goodluckxu-go/goapi/openapi"
)

type includeRouter struct {
	router      any
	prefix      string
	groupPrefix string
	isDocs      bool
	docsPath    string
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
	if !((value.Kind() == reflect.Ptr && value.Elem().Kind() == reflect.Struct) || value.Kind() == reflect.Struct) {
		err = fmt.Errorf("router must be a struct or struct pointer")
		return
	}
	pos := fmt.Sprintf("%v.%v", value.Type().PkgPath(), value.Type().Name())
	if value.Kind() == reflect.Ptr {
		pos = fmt.Sprintf("%v.(*%v)", value.Elem().Type().PkgPath(), value.Elem().Type().Name())
	}
	numMethod := value.NumMethod()
	for j := 0; j < numMethod; j++ {
		routerMethod := value.Method(j)
		// handle in param
		numIn := routerMethod.Type().NumIn()
		if numIn == 0 {
			continue
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
		pInfo := &pathInfo{
			pos:         fmt.Sprintf("%v.%v", pos, value.Type().Method(j).Name),
			value:       routerMethod,
			inTypes:     inTypes,
			middlewares: i.middlewares,
			isDocs:      i.isDocs,
			docsPath:    i.docsPath,
			groupPrefix: i.groupPrefix,
		}
		if len(pInfo.middlewares) > 0 {
			pInfo.pos += fmt.Sprintf(" (%v Middleware)", len(pInfo.middlewares))
		}
		numField := inputType.NumField()
		var in []*inParam
		for l := 0; l < numField; l++ {
			field := inputType.Field(l)
			switch field.Type {
			case reflect.TypeOf(Router{}):
				path := field.Tag.Get(tagPath)
				method := field.Tag.Get(tagMethod)
				if (i.prefix == "" && path == "") || method == "" {
					err = fmt.Errorf("the parameters must have a path and method present")
					return
				}
				methods := strings.Split(method, ",")
				for k, v := range methods {
					methods[k] = strings.ToUpper(v)
				}
				if !isMethod(methods) {
					err = fmt.Errorf("the method in the parameter does not exist %v", strings.Join(methods, ", "))
					return
				}
				paths := strings.Split(path, ",")
				for k, v := range paths {
					paths[k] = pathJoin(i.prefix, v)
				}
				pInfo.paths = paths
				pInfo.methods = methods
				pInfo.summary = field.Tag.Get(tagSummary)
				pInfo.desc = field.Tag.Get(tagDesc)
				pInfo.tags = tagStrs
				tag := field.Tag.Get(tagTags)
				if tag != "" {
					pInfo.tags = strings.Split(tag, ",")
				}
			default:
				if in, err = i.parseIn(field, []int{l}, ""); err != nil {
					return
				}
				for _, v := range in {
					for _, mediaType := range v.values.MediaTypes() {
						if mediaType.Tag() != "" {
							obj.mediaTypes[mediaType] = struct{}{}
						}
					}
				}
				pInfo.inParams = append(pInfo.inParams, in...)
			}
		}
		securityCount := 0
		for _, item := range pInfo.inParams {
			if inArray(item.inType, []InType{
				inTypeSecurityHTTPBearer,
				inTypeSecurityHTTPBearerJWT,
				inTypeSecurityHTTPBasic,
				inTypeSecurityApiKey,
			}) {
				securityCount++
			}
		}
		if securityCount > 0 {
			pInfo.pos += fmt.Sprintf(" (%v Security)", securityCount)
		}
		// handle out param
		numOut := routerMethod.Type().NumOut()
		if numOut > 1 {
			err = fmt.Errorf("method '%v' returns at most one result, has: %v", value.Type().Method(j).Name, numOut)
			return
		}
		if numOut == 1 {
			outType := routerMethod.Type().Out(0)
			pInfo.outParam = &outParam{
				structField: reflect.StructField{Type: outType},
			}
		}
		obj.paths = append(obj.paths, pInfo)
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
		params = append(childParams, params...)
	}
	return
}
