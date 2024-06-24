package goapi

import (
	"fmt"
	"github.com/goodluckxu-go/goapi/openapi"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
)

func newHandlerOpenAPI(api *API, paths []pathInfo, structs map[string]*structInfo) *handlerOpenAPI {
	return &handlerOpenAPI{
		api:   api,
		paths: paths,
		openapi: &openapi.OpenAPI{
			OpenAPI: openapi.Version,
			Info:    api.OpenAPIInfo,
			Servers: api.OpenAPIServers,
			Tags:    api.OpenAPITags,
		},
		structs:         structs,
		schemas:         map[string]map[string]*openapi.Schema{},
		relationSchemas: map[string][]string{},
		useSchemas:      map[string]struct{}{},
	}
}

type handlerOpenAPI struct {
	api             *API
	paths           []pathInfo
	openapi         *openapi.OpenAPI
	structs         map[string]*structInfo
	schemas         map[string]map[string]*openapi.Schema
	relationSchemas map[string][]string
	useSchemas      map[string]struct{}
}

func (h *handlerOpenAPI) Handle() *openapi.OpenAPI {
	h.handleStructs()
	h.handlePaths()
	h.handleUseSchemas()
	h.handleSchemas()
	if err := h.openapi.Validate(); err != nil {
		panic(err)
	}
	return h.openapi
}

func (h *handlerOpenAPI) handleUseSchemas() {
	handleSchemas := h.useSchemas
	for len(handleSchemas) > 0 {
		newHandleSchemas := map[string]struct{}{}
		for k, _ := range handleSchemas {
			for _, v := range h.relationSchemas[k] {
				if _, ok := h.useSchemas[v]; !ok {
					newHandleSchemas[v] = struct{}{}
				}
				h.useSchemas[v] = struct{}{}
			}
		}
		handleSchemas = newHandleSchemas
	}
}

func (h *handlerOpenAPI) handlePaths() {
	h.openapi.Paths = &openapi.Paths{}
	for _, path := range h.paths {
		if !path.isDocs {
			continue
		}
		h.setSecuritySchemes(path)
		pathItem := &openapi.PathItem{}
		if h.openapi.Paths.Value(path.path) != nil {
			pathItem = h.openapi.Paths.Value(path.path)
		}
		for _, method := range path.methods {
			operation := &openapi.Operation{}
			switch method {
			case http.MethodGet:
				if pathItem.Get != nil {
					operation = pathItem.Get
				}
			case http.MethodPut:
				if pathItem.Put != nil {
					operation = pathItem.Put
				}
			case http.MethodPost:
				if pathItem.Post != nil {
					operation = pathItem.Post
				}
			case http.MethodDelete:
				if pathItem.Delete != nil {
					operation = pathItem.Delete
				}
			case http.MethodOptions:
				if pathItem.Options != nil {
					operation = pathItem.Options
				}
			case http.MethodHead:
				if pathItem.Head != nil {
					operation = pathItem.Head
				}
			case http.MethodPatch:
				if pathItem.Patch != nil {
					operation = pathItem.Patch
				}
			case http.MethodTrace:
				if pathItem.Trace != nil {
					operation = pathItem.Trace
				}
			}
			h.setOperation(operation, path, method)
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
			h.openapi.Paths.Set(path.path, pathItem)
		}
	}
}

func (h *handlerOpenAPI) setOperation(operation *openapi.Operation, path pathInfo, method string) {
	operation.Summary = path.summary
	operation.Description = path.desc
	operation.Tags = path.tags
	operation.OperationId = strings.ToLower(strings.ReplaceAll(path.path+"_"+method, "/", "_"))
	var parameters []*openapi.Parameter
	var bodyMediaType MediaType
	bodyProperties := map[string]*openapi.Schema{}
	var bodyRequireds []string
	bodyContent := map[string]*openapi.MediaType{}
	var securityRequirements []*openapi.SecurityRequirement
	if len(h.api.responseMediaTypes) > 1 {
		securityRequirements = append(securityRequirements, &openapi.SecurityRequirement{
			"mediaType": []string{},
		})
	}
	for _, inputField := range path.inputFields {
		switch inputField.inType {
		case inTypePath, inTypeQuery, inTypeCookie, inTypeHeader:
			fType := h.convertType(inputField._type)
			childSchema := &openapi.Schema{
				Type:   []string{fType.typeStr},
				Format: fType.format,
			}
			childSchema.Enum = inputField.tag.enum
			childSchema.Default = inputField.tag._default
			switch fType.typeStr {
			case "integer", "number":
				childSchema.Maximum = inputField.tag.lte
				childSchema.ExclusiveMaximum = inputField.tag.lt
				childSchema.Minimum = inputField.tag.gte
				childSchema.ExclusiveMinimum = inputField.tag.gt
				childSchema.MultipleOf = inputField.tag.multiple
			case "string":
				childSchema.MaxLength = inputField.tag.max
				childSchema.MinLength = inputField.tag.min
				childSchema.Pattern = inputField.tag.regexp
			case "array":
				childSchema.MaxItems = inputField.tag.max
				childSchema.MinItems = inputField.tag.min
				childSchema.UniqueItems = inputField.tag.unique
				cfType := h.convertType(fType._type.Elem())
				childSchema.Items = &openapi.Schema{
					Type:   []string{cfType.typeStr},
					Format: cfType.format,
				}
			case "object":
				childSchema.MaxProperties = inputField.tag.max
				childSchema.MinProperties = inputField.tag.min
			}
			parameters = append(parameters, &openapi.Parameter{
				Name:        inputField.inTypeVal,
				In:          inputField.inType,
				Description: inputField.tag.desc,
				Schema:      childSchema,
				Required:    inputField.mediaTypes[0].required,
				Example:     inputField.tag.example,
			})
		case inTypeForm:
			if bodyMediaType != formMultipart {
				bodyMediaType = formUrlencoded
			}
			childSchema := &openapi.Schema{}
			h.setChildSchema(childSchema, inputField.deepTypes, inputField.inTypeVal)
			bodyProperties[inputField.inTypeVal] = childSchema
			if inputField.mediaTypes[0].required {
				bodyRequireds = append(bodyRequireds, inputField.inTypeVal)
			}
		case inTypeFile:
			bodyMediaType = formMultipart
			childSchema := &openapi.Schema{}
			h.setChildSchema(childSchema, inputField.deepTypes, inputField.inTypeVal)
			bodyProperties[inputField.inTypeVal] = childSchema
			if inputField.mediaTypes[0].required {
				bodyRequireds = append(bodyRequireds, inputField.inTypeVal)
			}
		case inTypeBody:
			lastType := inputField.deepTypes[len(inputField.deepTypes)-1]
			if lastType.isStruct {
				h.useSchemas[fmt.Sprintf("%v.%v", lastType._type.PkgPath(), lastType._type.Name())] = struct{}{}
			}
			for _, mediaType := range inputField.mediaTypes {
				childSchema := &openapi.Schema{}
				h.setChildSchema(childSchema, inputField.deepTypes, mediaType._type)
				bodyContent[string(typeToMediaTypeMap[mediaType._type])] = &openapi.MediaType{
					Schema: childSchema,
				}
			}
		case inTypeSecurityHTTPBearer, inTypeSecurityHTTPBasic, inTypeSecurityApiKey:
			securityRequirements = append(securityRequirements, &openapi.SecurityRequirement{
				inputField.name: []string{},
			})
		}
	}
	if len(bodyProperties) > 0 {
		bodyContent[string(bodyMediaType)] = &openapi.MediaType{
			Schema: &openapi.Schema{
				Type:       []string{"object"},
				Properties: bodyProperties,
				Required:   bodyRequireds,
			},
		}
	}
	if len(parameters) > 0 {
		operation.Parameters = parameters
	}
	if len(bodyContent) > 0 {
		operation.RequestBody = &openapi.RequestBody{
			Description: "Request Body",
			Content:     bodyContent,
		}
	}
	responses := &openapi.Responses{}
	if path.res != nil {
		resp := path.res
		lastType := resp.deepTypes[len(resp.deepTypes)-1]
		if lastType.isStruct {
			h.useSchemas[fmt.Sprintf("%v.%v", lastType._type.PkgPath(), lastType._type.Name())] = struct{}{}
		}
		responseContent := map[string]*openapi.MediaType{}
		for _, mediaType := range resp.mediaTypes {
			childSchema := &openapi.Schema{}
			h.setChildSchema(childSchema, resp.deepTypes, mediaType._type)
			responseContent[string(typeToMediaTypeMap[mediaType._type])] = &openapi.MediaType{
				Schema: childSchema,
			}
		}
		responses.Set("200", &openapi.Response{
			Description: "Successful Response",
			Content:     responseContent,
		})
	}
	if path.exceptRes != nil {
		resp := path.exceptRes
		lastType := resp.deepTypes[len(resp.deepTypes)-1]
		if lastType.isStruct {
			h.useSchemas[fmt.Sprintf("%v.%v", lastType._type.PkgPath(), lastType._type.Name())] = struct{}{}
		}
		responseContent := map[string]*openapi.MediaType{}
		for _, mediaType := range resp.mediaTypes {
			childSchema := &openapi.Schema{}
			h.setChildSchema(childSchema, resp.deepTypes, mediaType._type)
			responseContent[string(typeToMediaTypeMap[mediaType._type])] = &openapi.MediaType{
				Schema: childSchema,
			}
		}
		responses.Set("422", &openapi.Response{
			Description: "Validation Error",
			Content:     responseContent,
		})
	}
	operation.Responses = responses
	if len(securityRequirements) > 0 {
		operation.Security = securityRequirements
	}
}

func (h *handlerOpenAPI) setSecuritySchemes(path pathInfo) {
	securitySchemes := map[string]*openapi.SecurityScheme{}
	if len(h.api.responseMediaTypes) > 1 {
		securitySchemes["mediaType"] = &openapi.SecurityScheme{
			Type:        "apiKey",
			Name:        "media_type",
			In:          inTypeQuery,
			Description: "Switch the media type returned",
		}
	}
	if h.openapi.Components == nil {
		h.openapi.Components = &openapi.Components{}
	}
	if h.openapi.Components.SecuritySchemes != nil {
		securitySchemes = h.openapi.Components.SecuritySchemes
	}
	for _, inputFiled := range path.inputFields {
		switch inputFiled.inType {
		case inTypeSecurityHTTPBearer:
			securitySchemes[inputFiled.name] = &openapi.SecurityScheme{
				Type:   "http",
				Scheme: "bearer",
			}
		case inTypeSecurityHTTPBasic:
			securitySchemes[inputFiled.name] = &openapi.SecurityScheme{
				Type:   "http",
				Scheme: "basic",
			}
		case inTypeSecurityApiKey:
			securitySchemes[inputFiled.name] = &openapi.SecurityScheme{
				Type: "apiKey",
				Name: inputFiled.inTypeVal,
				In:   inputFiled.inTypeSecurity,
			}
		}
	}
	h.openapi.Components.SecuritySchemes = securitySchemes
}

func (h *handlerOpenAPI) handleSchemas() {
	if h.openapi.Components == nil {
		h.openapi.Components = &openapi.Components{}
	}

	for k, v := range h.structs {
		if strings.HasPrefix(k, prefixTempStruct) {
			continue
		}
		useKey := fmt.Sprintf("%v.%v", v.pkg, v.name)
		if _, ok := h.useSchemas[useKey]; !ok {
			continue
		}
		if h.schemas[v.openapiName] == nil {
			h.schemas[v.openapiName] = map[string]*openapi.Schema{}
		}
		properties, requiredMap := h.setStructSchema(v.fields)
		for key, val := range properties {
			h.schemas[v.openapiName][key] = &openapi.Schema{
				Type:       []string{"object"},
				Properties: val,
				Required:   requiredMap[key],
			}
		}
	}
	h.openapi.Components.Schemas = map[string]*openapi.Schema{}
	for key, schemas := range h.schemas {
		if len(schemas) == 1 {
			for _, v := range schemas {
				h.openapi.Components.Schemas[key] = v
			}
		} else {
			for k, v := range schemas {
				sKey := key
				if k != xmlType {
					sKey = key + "_" + k
				}
				h.openapi.Components.Schemas[sKey] = v
			}
		}
	}
}

func (h *handlerOpenAPI) setStructSchema(fields []fieldInfo) (properties map[string]map[string]*openapi.Schema, requiredMap map[string][]string) {
	properties = map[string]map[string]*openapi.Schema{}
	requiredMap = map[string][]string{}
	for _, v1 := range fields {
		for _, mediaType := range v1.mediaTypes {
			if mediaType.name == "-" {
				continue
			}
			childSchema := &openapi.Schema{}
			h.setChildSchema(childSchema, v1.deepTypes, mediaType._type)
			childSchema.Enum = v1.tag.enum
			childSchema.Default = v1.tag._default
			childSchema.Example = v1.tag.example
			childSchema.Description = v1.tag.desc
			fType := h.convertType(v1._type)
			switch fType.typeStr {
			case "integer", "number":
				childSchema.Maximum = v1.tag.lte
				childSchema.ExclusiveMaximum = v1.tag.lt
				childSchema.Minimum = v1.tag.gte
				childSchema.ExclusiveMinimum = v1.tag.gt
				childSchema.MultipleOf = v1.tag.multiple
			case "string":
				childSchema.MaxLength = v1.tag.max
				childSchema.MinLength = v1.tag.min
				childSchema.Pattern = v1.tag.regexp
			case "array":
				childSchema.MaxItems = v1.tag.max
				childSchema.MinItems = v1.tag.min
				childSchema.UniqueItems = v1.tag.unique
			case "object":
				childSchema.MaxProperties = v1.tag.max
				childSchema.MinProperties = v1.tag.min
			}
			if properties[mediaType._type] == nil {
				properties[mediaType._type] = map[string]*openapi.Schema{}
			}
			properties[mediaType._type][mediaType.name] = childSchema
			if mediaType.required {
				requiredMap[mediaType._type] = append(requiredMap[mediaType._type], mediaType.name)
			}
		}
	}
	return
}

func (h *handlerOpenAPI) handleStructs() {
	nameBaseCountMap := map[string]int{}
	nameExtCountMap := map[string]int{}
	for k, v := range h.structs {
		if strings.HasPrefix(k, prefixTempStruct) {
			continue
		}
		useKey := fmt.Sprintf("%v.%v", v.pkg, v.name)
		for _, field := range v.fields {
			if field._struct != nil {
				h.relationSchemas[useKey] = append(h.relationSchemas[useKey],
					fmt.Sprintf("%v.%v", field._struct.pkg, field._struct.name))
			}
		}
		nameBaseCountMap[k]++
		nameExtCountMap[filepath.Ext(k)[1:]]++
	}
	for k, v := range h.structs {
		if strings.HasPrefix(k, prefixTempStruct) {
			continue
		}
		v.openapiName = strings.Replace(k, "/", ".", -1)
		nameBase := filepath.Base(k)
		nameExt := filepath.Ext(k)[1:]
		if nameExtCountMap[nameExt] == 1 {
			v.openapiName = nameExt
		} else if nameBaseCountMap[nameBase] == 1 {
			v.openapiName = nameBase
		}
		h.structs[k] = v
	}
}

func (h *handlerOpenAPI) convertType(fType reflect.Type) (rs typeInfo) {
	rs._type = fType
	switch fType.Kind() {
	case reflect.Int, reflect.Uint:
		rs.typeStr = "integer"
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		rs.typeStr = "integer"
		rs.format = "int32"
	case reflect.Int64, reflect.Uint64:
		rs.typeStr = "integer"
		rs.format = "int64"
	case reflect.Float32:
		rs.typeStr = "number"
		rs.format = "float"
	case reflect.Float64:
		rs.typeStr = "number"
		rs.format = "double"
	case reflect.String:
		rs.typeStr = "string"
	case reflect.Bool:
		rs.typeStr = "boolean"
	case reflect.Slice:
		rs.typeStr = "array"
	case reflect.Map:
		rs.typeStr = "object"
	case reflect.Ptr:
		if fType == typeFile {
			rs.typeStr = "string"
			rs.format = "binary"
			return
		}
		rs = h.convertType(fType.Elem())
	case reflect.Struct:
		rs.typeStr = "object"
		rs.isStruct = true
	default:
		rs.typeStr = "string"
	}
	return
}

func (h *handlerOpenAPI) setChildSchema(schema *openapi.Schema, types []typeInfo, mediaType string) {
	if len(types) == 0 {
		return
	}
	tyInfo := h.convertType(types[0]._type)
	types = types[1:]
	schema.Type = []string{tyInfo.typeStr}
	schema.Format = tyInfo.format
	switch tyInfo._type.Kind() {
	case reflect.Map:
		childSchema := &openapi.Schema{}
		h.setChildSchema(childSchema, types, mediaType)
		schema.Properties = map[string]*openapi.Schema{
			"string": childSchema,
		}
	case reflect.Slice:
		childSchema := &openapi.Schema{}
		h.setChildSchema(childSchema, types, mediaType)
		schema.Items = childSchema
	case reflect.Struct:
		key := fmt.Sprintf("%v.%v", tyInfo._type.PkgPath(), tyInfo._type.Name())
		if key != "." {
			stInfo := h.structs[key]
			schemas := h.schemas[stInfo.openapiName]
			schemaKey := stInfo.openapiName
			if len(schemas) > 1 && mediaType != xmlType {
				schemaKey = stInfo.openapiName + "_" + mediaType
			}
			schema.Ref = "#/components/schemas/" + schemaKey
			return
		}
		key = fmt.Sprintf("%s%p", prefixTempStruct, tyInfo._type)
		stInfo := h.structs[key]
		properties, requiredMap := h.setStructSchema(stInfo.fields)
		schema.Required = requiredMap[mediaType]
		schema.Properties = properties[mediaType]
	default:
	}
}
