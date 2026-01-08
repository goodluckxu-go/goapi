package goapi

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/goodluckxu-go/goapi/openapi"
)

func newHandlerOpenAPI(api *API, handle *handler) *handlerOpenAPI {
	for _, openAPI := range handle.openapiMap {
		openAPI.OpenAPI = openapi.Version
	}
	return &handlerOpenAPI{
		api:        api,
		handle:     handle,
		schemas:    map[string]*openapi.Schema{},
		schemasMap: make(map[string]map[string]*openapi.Schema),
		sortRefMap: map[string]string{},
	}
}

type handlerOpenAPI struct {
	api        *API
	handle     *handler
	schemas    map[string]*openapi.Schema
	schemasMap map[string]map[string]*openapi.Schema
	sortRefMap map[string]string
}

func (h *handlerOpenAPI) Handle() map[string]*openapi.OpenAPI {
	h.handleStructs()
	h.handlePaths()
	h.handleUsePaths()
	for docsPath, schemas := range h.schemasMap {
		if h.handle.openapiMap[docsPath].Components == nil {
			h.handle.openapiMap[docsPath].Components = &openapi.Components{}
		}
		h.handle.openapiMap[docsPath].Components.Schemas = schemas
	}
	return h.handle.openapiMap
}

func (h *handlerOpenAPI) handleStructs() {
	for pkgName, stInfo := range h.handle.structs {
		for mediaType := range h.handle.mediaTypes {
			h.handleStruct(pkgName, stInfo, mediaType)
		}
	}
}

func (h *handlerOpenAPI) handleStruct(pkgName string, stInfo *structInfo, mediaType MediaType) {
	properties, required := h.handleParamFields(stInfo.fields, mediaType)
	schema := &openapi.Schema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
	refName := h.getOpenapiName(stInfo.openapiName, mediaType)
	h.sortRefMap[refName] = pkgName
	h.schemas[refName] = schema
}

func (h *handlerOpenAPI) handleParamFields(fields []*paramField, mediaType MediaType) (properties map[string]*openapi.Schema, required []string) {
	properties = map[string]*openapi.Schema{}
	for _, field := range fields {
		if field.anonymous {
			childStInfo := h.handle.structs[field.pkgName]
			childProperties, childRequired := h.handleParamFields(childStInfo.fields, mediaType)
			for k, v := range childProperties {
				properties[k] = v
			}
			required = append(required, childRequired...)
			continue
		}
		childSchema := &openapi.Schema{}
		name := h.handleParamField(childSchema, field, mediaType)
		properties[name.name] = childSchema
		if name.required {
			required = append(required, name.name)
		}
	}
	return
}

func (h *handlerOpenAPI) handleParamField(schema *openapi.Schema, field *paramField, mediaType MediaType) (name paramFieldName) {
	name = field.names.getFieldName(mediaType)
	kind := field.kind
	schema.Description = field.tag.desc
	schema.Deprecated = field.tag.deprecated
	if mediaType == XML {
		schema.XML = &openapi.XML{
			Extensions: map[string]any{},
		}
		for _, v := range name.split {
			switch v {
			case "attr":
				schema.XML.Attribute = true
			}
		}
	}
	if field.tag._default != nil {
		_default := field.tag._default
		isSet := h.handleNoJsonAndXmlExample(mediaType, &_default)
		if isSet {
			schema.Default = _default
		}
	}
	if field.tag.example != nil {
		example := field.tag.example
		isSet := h.handleNoJsonAndXmlExample(mediaType, &example)
		if isSet {
			schema.Examples = []any{example}
		}
	}
	switch name.inType {
	case inTypeFile:
		if field._type.ConvertibleTo(typeFile) {
			schema.Type = "string"
			schema.Format = "binary"
			return
		}
	case inTypeCookie:
		if field._type.ConvertibleTo(typeCookie) {
			schema.Type = "string"
			return
		}
	}
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
		if kind != reflect.Int {
			schema.Format = kind.String()
		}
		schema.Maximum = field.tag.lte
		schema.ExclusiveMaximum = field.tag.lt
		schema.Minimum = field.tag.gte
		schema.ExclusiveMinimum = field.tag.gt
		schema.MultipleOf = field.tag.multiple
		schema.Enum = field.tag.enum
	case reflect.Float32:
		schema.Type = "number"
		schema.Format = "float"
		schema.Maximum = field.tag.lte
		schema.ExclusiveMaximum = field.tag.lt
		schema.Minimum = field.tag.gte
		schema.ExclusiveMinimum = field.tag.gt
		schema.MultipleOf = field.tag.multiple
		schema.Enum = field.tag.enum
	case reflect.Float64:
		schema.Type = "number"
		schema.Format = "double"
		schema.Maximum = field.tag.lte
		schema.ExclusiveMaximum = field.tag.lt
		schema.Minimum = field.tag.gte
		schema.ExclusiveMinimum = field.tag.gt
		schema.MultipleOf = field.tag.multiple
		schema.Enum = field.tag.enum
	case reflect.String:
		schema.Type = "string"
		schema.MaxLength = field.tag.max
		schema.MinLength = field.tag.min
		schema.Pattern = field.tag.regexp
		schema.Enum = field.tag.enum
	case reflect.Bool:
		schema.Type = "boolean"
		schema.Enum = field.tag.enum
	case reflect.Array, reflect.Slice:
		if mediaType != "" && mediaType.IsStream() && field._type.ConvertibleTo(typeBytes) {
			schema.Type = "string"
			return
		}
		schema.Type = "array"
		schema.MaxItems = field.tag.max
		schema.MinItems = field.tag.min
		schema.UniqueItems = field.tag.unique
		childSchema := &openapi.Schema{}
		h.handleParamField(childSchema, field.fields[0], mediaType)
		schema.Items = childSchema
	case reflect.Map:
		schema.Type = "object"
		schema.MaxProperties = field.tag.max
		schema.MinProperties = field.tag.min
		childSchema := &openapi.Schema{}
		h.handleParamField(childSchema, field.fields[0], mediaType)
		schema.Properties = map[string]*openapi.Schema{
			field._type.Key().String(): childSchema,
		}
	case reflect.Struct:
		childStInfo := h.handle.structs[field.pkgName]
		schema.Ref = "#/components/schemas/" + h.getOpenapiName(childStInfo.openapiName, mediaType)
	case reflect.Ptr:
		if field._type.ConvertibleTo(typeCookie) || field._type.ConvertibleTo(typeFile) {
			schema.Type = "string"
		}
	default:
		schema.Type = "string"
	}
	return
}

func (h *handlerOpenAPI) handlePaths() {
	for _, path := range h.handle.paths {
		if path.inFs != nil {
			continue
		}

		h.handlePath(path)
	}
}

func (h *handlerOpenAPI) handleNoJsonAndXmlExample(mediaType MediaType, val *any) (isSet bool) {
	isSet = true
	if mediaType == JSON || val == nil {
		return
	}
	fType := reflect.TypeOf(*val)
	if isNormalType(fType) {
		return
	}
	// not a simple type without setting example, because it will be uniformly set using 'handleXmlExample'
	if mediaType == XML {
		isSet = false
		return
	}
	buf, err := mediaType.Marshaler(*val)
	if err != nil {
		*val = err.Error()
		return
	}
	fVal := reflect.ValueOf(val)
	_ = mediaType.Unmarshaler(io.NopCloser(bytes.NewBuffer(buf)), fVal)
	return
}

func (h *handlerOpenAPI) handleXmlExample(val *any) {
	if val == nil {
		return
	}
	fVal := reflect.ValueOf(val)
	if isNormalType(fVal.Elem().Type()) {
		return
	}
	buf, err := xml.MarshalIndent(*val, "", "	")
	if err != nil {
		*val = xml.Header + err.Error()
		return
	}
	*val = xml.Header + string(buf)
}

func (h *handlerOpenAPI) handleUsePaths() {
	for _, path := range h.handle.paths {
		if path.inFs != nil {
			continue
		}
		if h.schemasMap[path.docsPath] == nil {
			h.schemasMap[path.docsPath] = map[string]*openapi.Schema{}
		}
		openAPI := h.handle.openapiMap[path.docsPath]
		for _, p := range path.paths {
			if !path.isDocs {
				continue
			}
			setPath, _, _ := h.getMatchAllPath(p)
			pathItem := openAPI.Paths.Value(setPath)
			h.handleUseOperation(path, pathItem.Get)
			h.handleUseOperation(path, pathItem.Put)
			h.handleUseOperation(path, pathItem.Post)
			h.handleUseOperation(path, pathItem.Delete)
			h.handleUseOperation(path, pathItem.Options)
			h.handleUseOperation(path, pathItem.Head)
			h.handleUseOperation(path, pathItem.Patch)
			h.handleUseOperation(path, pathItem.Trace)
		}

	}
}

func (h *handlerOpenAPI) handleUseOperation(path *pathInfo, operation *openapi.Operation) {
	if operation == nil {
		return
	}
	if operation.RequestBody != nil {
		for _, content := range operation.RequestBody.Content {
			h.handleUseSchema(path.docsPath, content.Schema)
		}
	}
	if operation.Responses != nil {
		for _, response := range operation.Responses.Responses() {
			if response != nil {
				for _, content := range response.Content {
					h.handleUseSchema(path.docsPath, content.Schema)
				}
			}
		}
	}
}

func (h *handlerOpenAPI) handleUseSchema(docsPath string, scheme *openapi.Schema) {
	if scheme == nil {
		return
	}
	if scheme.Ref == "" {
		switch scheme.Type {
		case "object":
			for _, childSchema := range scheme.Properties {
				h.handleUseSchema(docsPath, childSchema)
			}
		case "array":
			h.handleUseSchema(docsPath, scheme.Items)
		}
		return
	}
	refName, mediaType := h.getMediaTypeByRef(scheme.Ref)
	h.handleDependsStruct(docsPath, h.sortRefMap[refName], mediaType)
}

func (h *handlerOpenAPI) getMediaTypeByRef(ref string) (refName string, mediaType MediaType) {
	refName = strings.TrimPrefix(ref, "#/components/schemas/")
	lastIdx := strings.LastIndex(refName, ".")
	for mType := range h.handle.mediaTypes {
		mediaType = mType
		break
	}
	if lastIdx != -1 && len(h.handle.mediaTypes) > 1 {
		for mType := range h.handle.mediaTypes {
			if mType.Tag() == refName[lastIdx+1:] {
				mediaType = mType
				break
			}
		}
	}
	return
}

func (h *handlerOpenAPI) handlePath(path *pathInfo) {
	openAPI := h.handle.openapiMap[path.docsPath]
	for _, p := range path.paths {
		if !path.isDocs {
			continue
		}
		h.handleSecuritySchemes(openAPI, path)
		setPath, pathName, isMatchAll := h.getMatchAllPath(p)
		if openAPI.Paths == nil {
			openAPI.Paths = &openapi.Paths{}
		}
		pathItem := openAPI.Paths.Value(setPath)
		if pathItem == nil {
			pathItem = &openapi.PathItem{}
		}
		for _, method := range path.methods {
			operation := &openapi.Operation{}
			switch method {
			case http.MethodGet:
				pathItem.Get = operation
			case http.MethodPut:
				pathItem.Put = operation
			case http.MethodPost:
				pathItem.Post = operation
			case http.MethodDelete:
				pathItem.Delete = operation
			case http.MethodOptions:
				pathItem.Options = operation
			case http.MethodHead:
				pathItem.Head = operation
			case http.MethodPatch:
				pathItem.Patch = operation
			case http.MethodTrace:
				pathItem.Trace = operation
			}
			operation.OperationId = fmt.Sprintf("%v%v", strings.ToLower(method), strings.ReplaceAll(setPath, "/", "_"))
			h.handleOperation(operation, path, setPath, pathName, isMatchAll)
		}
		openAPI.Paths.Set(setPath, pathItem)
	}
}

func (h *handlerOpenAPI) handleSecuritySchemes(openAPI *openapi.OpenAPI, path *pathInfo) {
	securitySchemes := map[string]*openapi.SecurityScheme{}
	if openAPI.Components != nil && openAPI.Components.SecuritySchemes != nil {
		securitySchemes = openAPI.Components.SecuritySchemes
	}
	if len(h.api.responseMediaTypes) > 1 {
		securitySchemes["mediaType"] = &openapi.SecurityScheme{
			Type:        "apiKey",
			Name:        returnMediaTypeField,
			In:          inTypeQuery.Tag(),
			Description: "Switch the media type returned",
		}
	}
	for _, in := range path.inParams {
		switch in.inType {
		case inTypePath, inTypeQuery, inTypeHeader, inTypeCookie:
			if in.parentInType == "" {
				continue
			}
			securitySchemes[in.structField.Name] = &openapi.SecurityScheme{
				Type:        "apiKey",
				Name:        in.values[0].name,
				In:          in.inType.Tag(),
				Description: in.field.tag.desc,
			}
		case inTypeSecurityHTTPBearer:
			securitySchemes[in.structField.Name] = &openapi.SecurityScheme{
				Type:   "http",
				Scheme: "bearer",
			}
		case inTypeSecurityHTTPBearerJWT:
			securitySchemes[in.structField.Name] = &openapi.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			}
		case inTypeSecurityHTTPBasic:
			securitySchemes[in.structField.Name] = &openapi.SecurityScheme{
				Type:   "http",
				Scheme: "basic",
			}
		}
	}
	if len(securitySchemes) > 0 {
		if openAPI.Components == nil {
			openAPI.Components = &openapi.Components{}
		}
		openAPI.Components.SecuritySchemes = securitySchemes
	}
}

func (h *handlerOpenAPI) handleOperation(operation *openapi.Operation, path *pathInfo, setPath, pathName string, isMatchAll bool) {
	operation.Tags = path.tags
	operation.Summary = path.summary
	operation.Description = path.desc
	bodyContentMap := map[string]*openapi.MediaType{}
	bodyProperties := map[string]*openapi.Schema{}
	var bodyMediaType MediaType
	var bodyRequired []string
	var securityRequirements []*openapi.SecurityRequirement
	if len(h.api.responseMediaTypes) > 1 {
		securityRequirements = append(securityRequirements, &openapi.SecurityRequirement{
			"mediaType": []string{},
		})
	}
	bodyDesc := ""
	for _, in := range path.inParams {
		switch in.inType {
		case inTypePath, inTypeQuery, inTypeHeader, inTypeCookie:
			if in.parentInType != "" {
				securityRequirements = append(securityRequirements, &openapi.SecurityRequirement{
					in.structField.Name: []string{},
				})
				continue
			}
			name := in.values[0]
			if in.inType == inTypePath && !h.isParamPath(name.name, setPath) {
				continue
			}
			schema := &openapi.Schema{}
			h.handleParamField(schema, in.field, "")
			parameter := &openapi.Parameter{
				Name:        name.name,
				In:          in.inType.Tag(),
				Description: in.field.tag.desc,
				Required:    name.required,
				Deprecated:  in.field.tag.deprecated,
				Schema:      schema,
				Example:     in.field.tag.example,
			}
			if in.inType == inTypePath && pathName == name.name && isMatchAll {
				parameter.Extensions = map[string]any{
					"x-match": "*",
				}
			}
			if in.inType == inTypeCookie {
				parameter.Example = "document.cookie"
			}
			operation.Parameters = append(operation.Parameters, parameter)
		case inTypeForm:
			name := in.values[0]
			if bodyMediaType != formMultipart {
				bodyMediaType = formUrlencoded
			}
			schema := &openapi.Schema{}
			h.handleParamField(schema, in.field, "")
			bodyProperties[name.name] = schema
			if name.required {
				bodyRequired = append(bodyRequired, name.name)
			}
		case inTypeFile:
			name := in.values[0]
			bodyMediaType = formMultipart
			schema := &openapi.Schema{}
			h.handleParamField(schema, in.field, "")
			bodyProperties[name.name] = schema
			if name.required {
				bodyRequired = append(bodyRequired, name.name)
			}
		case inTypeBody:
			bodyDesc = in.field.tag.desc
			for _, value := range in.values {
				schema := &openapi.Schema{}
				h.handleParamField(schema, in.field, value.mediaType)
				if in.example != nil && value.mediaType == XML {
					example := in.example
					h.handleXmlExample(&example)
					schema.Examples = []any{example}
				}
				if value.mediaType == XML {
					schema.XML = &openapi.XML{
						Name: in.field._type.Name(),
					}
				}
				bodyContentMap[string(value.mediaType)] = &openapi.MediaType{
					Schema: schema,
				}
			}
		case inTypeSecurityHTTPBearer, inTypeSecurityHTTPBearerJWT, inTypeSecurityHTTPBasic:
			securityRequirements = append(securityRequirements, &openapi.SecurityRequirement{
				in.structField.Name: []string{},
			})
		}
	}

	if len(bodyProperties) > 0 {
		bodyContentMap[string(bodyMediaType)] = &openapi.MediaType{
			Schema: &openapi.Schema{
				Type:       "object",
				Properties: bodyProperties,
				Required:   bodyRequired,
			},
		}
	}
	if len(bodyContentMap) > 0 {
		operation.RequestBody = &openapi.RequestBody{
			Description: bodyDesc,
			Content:     bodyContentMap,
		}
	}
	resMap := map[string]*openapi.Response{}
	if path.outParam != nil {
		resContentMap := map[string]*openapi.MediaType{}
		contentType := path.outParam.httpHeader.Get("Content-Type")
		if contentType == "" {
			for _, mediaType := range h.api.responseMediaTypes {
				schema := &openapi.Schema{}
				if mediaType.IsStream() {
					schema.Type = "string"
				} else {
					h.handleParamField(schema, path.outParam.field, mediaType)
					if mediaType == XML {
						schema.XML = &openapi.XML{
							Name: path.outParam.field._type.Name(),
						}
					}
					if path.outParam.example != nil && mediaType == XML {
						example := path.outParam.example
						h.handleXmlExample(&example)
						schema.Examples = []any{example}
					}
				}
				resContentMap[string(mediaType)] = &openapi.MediaType{
					Schema: schema,
				}
			}
		} else {
			mediaType := MediaType(contentType)
			schema := &openapi.Schema{}
			if mediaType.IsStream() {
				schema.Type = "string"
			} else {
				h.handleParamField(schema, path.outParam.field, mediaType)
				if mediaType == XML {
					schema.XML = &openapi.XML{
						Name: path.outParam.field._type.Name(),
					}
				}
				if path.outParam.example != nil && mediaType == XML {
					example := path.outParam.example
					h.handleXmlExample(&example)
					schema.Examples = []any{example}
				}
			}
			resContentMap[string(mediaType)] = &openapi.MediaType{
				Schema: schema,
			}
		}
		header := map[string]*openapi.Header{}
		for key, head := range path.outParam.httpHeader {
			if key == "Content-Type" {
				continue
			}
			header[key] = &openapi.Header{
				Description: strings.Join(head, ", "),
			}
		}
		resMap[toString(path.outParam.httpStatus)] = &openapi.Response{
			Description: "Successful Response",
			Content:     resContentMap,
			Headers:     header,
		}
	}
	if h.handle.except != nil {
		resContentMap := map[string]*openapi.MediaType{}
		contentType := h.handle.except.httpHeader.Get("Content-Type")
		if contentType == "" {
			for _, mediaType := range h.api.responseMediaTypes {
				schema := &openapi.Schema{}
				if mediaType.IsStream() {
					schema.Type = "string"
				} else {
					h.handleParamField(schema, h.handle.except.field, mediaType)
					if mediaType == XML {
						schema.XML = &openapi.XML{
							Name: h.handle.except.field._type.Name(),
						}
					}
					if h.handle.except.example != nil && mediaType == XML {
						example := h.handle.except.example
						h.handleXmlExample(&example)
						schema.Examples = []any{example}
					}
				}
				resContentMap[string(mediaType)] = &openapi.MediaType{
					Schema: schema,
				}
			}
		} else {
			mediaType := MediaType(contentType)
			schema := &openapi.Schema{}
			if mediaType.IsStream() {
				schema.Type = "string"
			} else {
				h.handleParamField(schema, h.handle.except.field, mediaType)
				if mediaType == XML {
					schema.XML = &openapi.XML{
						Name: h.handle.except.field._type.Name(),
					}
				}
				if h.handle.except.example != nil && mediaType == XML {
					example := h.handle.except.example
					h.handleXmlExample(&example)
					schema.Examples = []any{example}
				}
			}
			resContentMap[string(mediaType)] = &openapi.MediaType{
				Schema: schema,
			}
		}
		header := map[string]*openapi.Header{}
		for key, head := range h.handle.except.httpHeader {
			if key == "Content-Type" {
				continue
			}
			header[key] = &openapi.Header{
				Description: strings.Join(head, ", "),
			}
		}
		resMap["422"] = &openapi.Response{
			Description: "Validation Error",
			Content:     resContentMap,
			Headers:     header,
		}
	}
	if len(resMap) > 0 {
		operation.Responses = &openapi.Responses{}
		for status, res := range resMap {
			operation.Responses.Set(status, res)
		}
	}
	if len(securityRequirements) > 0 {
		operation.Security = securityRequirements
	}
}

func (h *handlerOpenAPI) isParamPath(name string, setPath string) (ok bool) {
	return strings.Contains(setPath, "{"+name+"}")
}

func (h *handlerOpenAPI) getMatchAllPath(fullPath string) (setPath, pathName string, isMatchAll bool) {
	setPath = fullPath
	lastIdx := len(fullPath) - 1
	if fullPath[lastIdx] != '}' {
		return
	}
	startIdx := lastIdx - 1
	for ; startIdx > 0 && fullPath[startIdx] != '{'; startIdx-- {
	}
	if fullPath[startIdx] != '{' {
		return
	}
	tmp := fullPath[startIdx+1 : lastIdx]
	pathName = strings.TrimSuffix(tmp, ":*")
	if tmp != pathName {
		isMatchAll = true
	}
	setPath = fullPath[:startIdx] + "{" + pathName + "}"
	return
}

func (h *handlerOpenAPI) handleDependsStruct(docsPath string, pkgName string, mediaType MediaType) {
	if h.schemasMap[docsPath] == nil {
		h.schemasMap[docsPath] = make(map[string]*openapi.Schema)
	}
	openapiName := h.getOpenapiNameByPkgName(pkgName, mediaType)
	if _, ok := h.schemasMap[docsPath][openapiName]; ok {
		return
	}
	h.schemasMap[docsPath][openapiName] = h.schemas[openapiName]
	list := h.handle.structDepends[pkgName]
	for _, val := range list {
		h.handleDependsStruct(docsPath, val, mediaType)
	}
}

func (h *handlerOpenAPI) getOpenapiNameByPkgName(pkgName string, mediaType MediaType) string {
	return h.getOpenapiName(h.handle.structs[pkgName].openapiName, mediaType)
}

func (h *handlerOpenAPI) getOpenapiName(openapiName string, mediaType MediaType) string {
	if len(h.handle.mediaTypes) < 2 {
		return openapiName
	}
	return fmt.Sprintf("%v.%v", openapiName, mediaType.Tag())
}
