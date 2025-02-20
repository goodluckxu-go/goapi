package goapi

import (
	"fmt"
	"github.com/goodluckxu-go/goapi/openapi"
	"log"
	"math"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

func newHandlerOpenAPI(api *API, handle *handler) *handlerOpenAPI {
	openapiMap := map[string]*openapi.OpenAPI{}
	if len(handle.openapiSetMap) > 0 {
		for k, v := range handle.openapiSetMap {
			v.OpenAPI = openapi.Version
			openapiMap[k] = v
		}
	} else {
		openapiMap[api.docsPath] = &openapi.OpenAPI{
			OpenAPI: openapi.Version,
			Info:    api.OpenAPIInfo,
			Servers: api.OpenAPIServers,
			Tags:    api.OpenAPITags,
		}
	}
	handleApi := &handlerOpenAPI{
		api:             api,
		handle:          handle,
		openapiMap:      openapiMap,
		schemas:         map[string]map[MediaType]*openapi.Schema{},
		isMullMediaType: len(handle.allMediaTypes) > 1,
	}
	for k := range handle.allMediaTypes {
		handleApi.singleMediaType = k
	}
	return handleApi
}

type handlerOpenAPI struct {
	api             *API
	handle          *handler
	openapiMap      map[string]*openapi.OpenAPI
	schemas         map[string]map[MediaType]*openapi.Schema
	isMullMediaType bool
	singleMediaType MediaType
}

func (h *handlerOpenAPI) Handle() map[string]*openapi.OpenAPI {
	h.handleStructs()
	h.handlePaths()
	h.handleSchemas()
	h.handleUseSchemas()
	for _, oApi := range h.openapiMap {
		if err := oApi.Validate(); err != nil {
			log.Fatal(err)
		}
	}
	return h.openapiMap
}

func (h *handlerOpenAPI) handleUseSchemas() {
	for _, oApi := range h.openapiMap {
		isDel := true
		for isDel {
			buf, _ := oApi.MarshalJSON()
			str := string(buf)
			isDel = false
			for k := range oApi.Components.Schemas {
				ref := "#/components/schemas/" + k
				if !strings.Contains(str, ref) {
					delete(oApi.Components.Schemas, k)
					isDel = true
				}
			}
		}
	}

}

func (h *handlerOpenAPI) handlePaths() {
	for _, path := range h.handle.paths {
		if !path.isDocs {
			continue
		}
		if h.openapiMap[path.docsPath] == nil {
			continue
		}
		if h.openapiMap[path.docsPath].Paths == nil {
			h.openapiMap[path.docsPath].Paths = &openapi.Paths{}
		}
		h.setSecuritySchemes(path)
		pathItem := &openapi.PathItem{}
		if h.openapiMap[path.docsPath].Paths.Value(path.path) != nil {
			pathItem = h.openapiMap[path.docsPath].Paths.Value(path.path)
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
			h.openapiMap[path.docsPath].Paths.Set(path.path, pathItem)
		}
	}
}

func (h *handlerOpenAPI) setOperation(operation *openapi.Operation, path *pathInfo, method string) {
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
	bodyDesc := "Request Body"
	for _, inputField := range path.inputFields {
		switch inputField.inType {
		case inTypePath, inTypeQuery, inTypeCookie, inTypeHeader:
			fType := h.convertType(inputField._type, false)
			h.mergeTag(inputField.tag, fType)
			childSchema := &openapi.Schema{
				Type:   fType.typeStr,
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
				cfType := h.convertType(fType._type.Elem(), false)
				childSchema.Items = &openapi.Schema{
					Type:    cfType.typeStr,
					Format:  cfType.format,
					Maximum: cfType.lte,
					Minimum: cfType.gte,
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
				Required:    inputField.required,
				Example:     inputField.tag.example,
			})
		case inTypeForm:
			if bodyMediaType != formMultipart {
				bodyMediaType = formUrlencoded
			}
			childSchema := &openapi.Schema{}
			h.setChildSchema(childSchema, inputField.deepTypes, "", false)
			bodyProperties[inputField.inTypeVal] = childSchema
			if inputField.required {
				bodyRequireds = append(bodyRequireds, inputField.inTypeVal)
			}
		case inTypeFile:
			bodyMediaType = formMultipart
			childSchema := &openapi.Schema{}
			h.setChildSchema(childSchema, inputField.deepTypes, "", false)
			bodyProperties[inputField.inTypeVal] = childSchema
			if inputField.required {
				bodyRequireds = append(bodyRequireds, inputField.inTypeVal)
			}
		case inTypeBody:
			lastType := inputField.deepTypes[len(inputField.deepTypes)-1]
			if inputField.tag.desc != "" {
				bodyDesc = inputField.tag.desc
			}
			for _, mediaType := range inputField.mediaTypes {
				childSchema := &openapi.Schema{}
				isBodyNotJsonXml := false
				if mediaType != JSON && mediaType != XML {
					isBodyNotJsonXml = true
				}
				h.setChildSchema(childSchema, inputField.deepTypes, mediaType, isBodyNotJsonXml)
				if mediaType == XML {
					childSchema.XML = &openapi.XML{
						Name: h.getStructBaseName(lastType._type.Name()),
					}
				}
				bodyContent[string(mediaType)] = &openapi.MediaType{
					Schema: childSchema,
				}
			}
		case inTypeSecurityHTTPBearer, inTypeSecurityHTTPBearerJWT, inTypeSecurityHTTPBasic, inTypeSecurityApiKey:
			securityRequirements = append(securityRequirements, &openapi.SecurityRequirement{
				inputField.name: []string{},
			})
		}
	}
	if len(bodyProperties) > 0 {
		bodyContent[string(bodyMediaType)] = &openapi.MediaType{
			Schema: &openapi.Schema{
				Type:       "object",
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
			Description: bodyDesc,
			Content:     bodyContent,
		}
	}
	responses := &openapi.Responses{}
	if path.res != nil {
		resp := path.res
		lastType := resp.deepTypes[len(resp.deepTypes)-1]
		responseContent := map[string]*openapi.MediaType{}
		for _, mediaType := range resp.mediaTypes {
			childSchema := &openapi.Schema{}
			isBodyNotJsonXml := false
			if mediaType != JSON && mediaType != XML {
				isBodyNotJsonXml = true
			}
			h.setChildSchema(childSchema, resp.deepTypes, mediaType, isBodyNotJsonXml)
			if mediaType == XML {
				childSchema.XML = &openapi.XML{
					Name: h.getStructBaseName(lastType._type.Name()),
				}
			}
			responseContent[string(mediaType)] = &openapi.MediaType{
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
		responseContent := map[string]*openapi.MediaType{}
		for _, mediaType := range resp.mediaTypes {
			childSchema := &openapi.Schema{}
			isBodyNotJsonXml := false
			if mediaType != JSON && mediaType != XML {
				isBodyNotJsonXml = true
			}
			h.setChildSchema(childSchema, resp.deepTypes, mediaType, isBodyNotJsonXml)
			if mediaType == XML {
				childSchema.XML = &openapi.XML{
					Name: h.getStructBaseName(lastType._type.Name()),
				}
			}
			responseContent[string(mediaType)] = &openapi.MediaType{
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

func (h *handlerOpenAPI) setSecuritySchemes(path *pathInfo) {
	if h.openapiMap[path.docsPath] == nil {
		return
	}
	securitySchemes := map[string]*openapi.SecurityScheme{}
	if len(h.api.responseMediaTypes) > 1 {
		securitySchemes["mediaType"] = &openapi.SecurityScheme{
			Type:        "apiKey",
			Name:        "media_type",
			In:          inTypeQuery,
			Description: "Switch the media type returned",
		}
	}
	if h.openapiMap[path.docsPath].Components == nil {
		h.openapiMap[path.docsPath].Components = &openapi.Components{}
	}
	if h.openapiMap[path.docsPath].Components.SecuritySchemes != nil {
		securitySchemes = h.openapiMap[path.docsPath].Components.SecuritySchemes
	}
	for _, inputFiled := range path.inputFields {
		switch inputFiled.inType {
		case inTypeSecurityHTTPBearer:
			securitySchemes[inputFiled.name] = &openapi.SecurityScheme{
				Type:        "http",
				Scheme:      "bearer",
				Description: inputFiled.tag.desc,
			}
		case inTypeSecurityHTTPBearerJWT:
			securitySchemes[inputFiled.name] = &openapi.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  inputFiled.tag.desc,
			}
		case inTypeSecurityHTTPBasic:
			securitySchemes[inputFiled.name] = &openapi.SecurityScheme{
				Type:        "http",
				Scheme:      "basic",
				Description: inputFiled.tag.desc,
			}
		case inTypeSecurityApiKey:
			securitySchemes[inputFiled.name] = &openapi.SecurityScheme{
				Type:        "apiKey",
				Name:        inputFiled.inTypeVal,
				In:          inputFiled.inTypeSecurity,
				Description: inputFiled.tag.desc,
			}
		}
	}
	if len(securitySchemes) > 0 {
		h.openapiMap[path.docsPath].Components.SecuritySchemes = securitySchemes
	}
}

func (h *handlerOpenAPI) handleSchemas() {
	for k, v := range h.handle.structs {
		if strings.HasPrefix(k, prefixTempStruct) {
			continue
		}
		if h.schemas[v.openapiName] == nil {
			h.schemas[v.openapiName] = map[MediaType]*openapi.Schema{}
		}
		properties, requiredMap := h.setStructSchema(v.fields)
		for key, val := range properties {
			h.schemas[v.openapiName][key] = &openapi.Schema{
				Type:       "object",
				Properties: val,
				Required:   requiredMap[key],
			}
		}
	}
	allSchemas := map[string]*openapi.Schema{}
	for key, schemas := range h.schemas {
		if !h.isMullMediaType {
			allSchemas[key] = schemas[h.singleMediaType]
		} else {
			for k, v := range schemas {
				allSchemas[key+"_"+mediaTypeToTypeMap[k]] = v
			}
		}
	}
	for _, oApi := range h.openapiMap {
		schemas := map[string]*openapi.Schema{}
		for k, v := range allSchemas {
			tmp := v
			schemas[k] = tmp
		}
		if oApi.Components == nil {
			oApi.Components = &openapi.Components{}
		}
		oApi.Components.Schemas = schemas
	}
}

func (h *handlerOpenAPI) setStructSchema(fields []fieldInfo) (properties map[MediaType]map[string]*openapi.Schema, requiredMap map[MediaType][]string) {
	properties = map[MediaType]map[string]*openapi.Schema{}
	requiredMap = map[MediaType][]string{}
	for _, v1 := range fields {
		for mType, fInfo := range v1.fieldMap {
			if fInfo.name == "-" {
				continue
			}
			childSchema := &openapi.Schema{}
			h.setChildSchema(childSchema, v1.deepTypes, mType, false)
			childSchema.Enum = v1.tag.enum
			childSchema.Default = v1.tag._default
			if v1.tag.example != nil {
				childSchema.Examples = []any{v1.tag.example}
			}
			childSchema.Description = v1.tag.desc
			fType := h.convertType(v1._type, false)
			h.mergeTag(v1.tag, fType)
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
				cfType := h.convertType(fType._type.Elem(), false)
				if childSchema.Items == nil {
					childSchema.Items = &openapi.Schema{}
				}
				childSchema.Items.Type = cfType.typeStr
				childSchema.Items.Format = cfType.format
				childSchema.Items.Maximum = cfType.lte
				childSchema.Items.Minimum = cfType.gte
			case "object":
				childSchema.MaxProperties = v1.tag.max
				childSchema.MinProperties = v1.tag.min
			}
			// xml
			if mType == XML && fInfo.xml != nil {
				if fInfo.xml.attr {
					childSchema.XML = &openapi.XML{
						Attribute: true,
					}
				}
				if fInfo.xml.innerxml {
					continue
				}
				if len(fInfo.xml.childs) > 0 {
					childSchema.Type = "object"
					h.setXmlChildList(childSchema, fInfo.xml.childs)
				}
			}
			if properties[mType] == nil {
				properties[mType] = map[string]*openapi.Schema{}
			}
			properties[mType][fInfo.name] = childSchema
			if fInfo.required {
				requiredMap[mType] = append(requiredMap[mType], fInfo.name)
			}
		}
	}
	return
}

func (h *handlerOpenAPI) handleStructs() {
	nameMap := map[string]map[string]struct{}{}
	nameBaseMap := map[string]map[string]struct{}{}
	for k := range h.handle.structs {
		if strings.HasPrefix(k, prefixTempStruct) {
			continue
		}
		pkg, name, baseName := h.parseOpenapiName(k)
		if nameMap[name] == nil {
			nameMap[name] = map[string]struct{}{}
		}
		nameMap[name][pkg] = struct{}{}
		if nameBaseMap[baseName] == nil {
			nameBaseMap[baseName] = map[string]struct{}{}
		}
		nameBaseMap[baseName][name] = struct{}{}
	}
	for k, v := range h.handle.structs {
		if strings.HasPrefix(k, prefixTempStruct) {
			continue
		}
		v.openapiName = strings.Replace(k, "/", ".", -1)
		_, name, baseName := h.parseOpenapiName(k)
		if len(nameBaseMap[baseName]) == 1 {
			v.openapiName = baseName
		} else if len(nameMap[name]) == 1 {
			v.openapiName = name
		}
		h.handle.structs[k] = v
	}
}

func (h *handlerOpenAPI) parseOpenapiName(s string) (pkg, name, baseName string) {
	idx := strings.Index(s, ".")
	if idx == -1 {
		name = s
		return
	}
	pkg = s[:idx]
	name = s[idx+1:]
	baseName = name
	lIdx := strings.Index(name, "[")
	rIdx := strings.LastIndex(name, "]")
	if lIdx != -1 && rIdx != -1 && lIdx < rIdx {
		baseName = name[:lIdx]
		name = regexp.MustCompile(`(\w+\[)\w+\.(.*?)`).ReplaceAllString(name, "$1$2")
	}
	return
}

func (h *handlerOpenAPI) convertType(fType reflect.Type, isBodyNotJsonXml bool) (rs typeInfo) {
	rs._type = fType
	switch fType.Kind() {
	case reflect.Int:
		rs.typeStr = "integer"
	case reflect.Int8:
		rs.typeStr = "integer"
		rs.format = "int8"
		rs.gte = toPtr(float64(math.MinInt8))
		rs.lte = toPtr(float64(math.MaxInt8))
	case reflect.Int16:
		rs.typeStr = "integer"
		rs.format = "int16"
		rs.gte = toPtr(float64(math.MinInt16))
		rs.lte = toPtr(float64(math.MaxInt16))
	case reflect.Uint8:
		rs.typeStr = "integer"
		rs.format = "uint8"
		rs.gte = toPtr(float64(0))
		rs.lte = toPtr(float64(math.MaxUint8))
	case reflect.Uint16:
		rs.typeStr = "integer"
		rs.format = "uint16"
		rs.gte = toPtr(float64(0))
		rs.lte = toPtr(float64(math.MaxUint16))
	case reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint32, reflect.Uint64:
		rs.typeStr = "integer"
		rs.format = fType.Kind().String()
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
		if fType == typeBytes && isBodyNotJsonXml {
			rs.typeStr = "string"
			return
		}
		rs.typeStr = "array"
	case reflect.Map:
		rs.typeStr = "object"
	case reflect.Ptr:
		switch fType {
		case typeFile:
			rs.typeStr = "string"
			rs.format = "binary"
			return
		case typeCookie:
			rs.typeStr = "string"
			return
		}
		if fType.Implements(interfaceIoReadCloser) {
			rs.typeStr = "string"
			rs.format = "binary"
			return
		}
		rs = h.convertType(fType.Elem(), isBodyNotJsonXml)
	case reflect.Struct:
		rs.typeStr = "object"
		rs.isStruct = true
	default:
		rs.typeStr = "string"
	}
	return
}

func (h *handlerOpenAPI) setXmlChildList(schema *openapi.Schema, childList []string) {
	child := childList[0]
	childList = childList[1:]
	childSchema := &openapi.Schema{
		Type: "string",
	}
	if len(childList) > 0 {
		childSchema.Type = "object"
		h.setXmlChildList(childSchema, childList)
	}
	schema.Properties = map[string]*openapi.Schema{
		child: childSchema,
	}
}

func (h *handlerOpenAPI) setChildSchema(schema *openapi.Schema, types []typeInfo, mediaType MediaType, isBodyNotJsonXml bool) {
	if len(types) == 0 {
		return
	}
	tyInfo := h.convertType(types[0]._type, isBodyNotJsonXml)
	types = types[1:]
	schema.Type = tyInfo.typeStr
	schema.Format = tyInfo.format
	switch tyInfo._type.Kind() {
	case reflect.Map:
		childSchema := &openapi.Schema{}
		h.setChildSchema(childSchema, types, mediaType, isBodyNotJsonXml)
		schema.Properties = map[string]*openapi.Schema{
			"string": childSchema,
		}
	case reflect.Slice:
		if tyInfo._type == typeBytes && isBodyNotJsonXml {
			return
		}
		childSchema := &openapi.Schema{}
		h.setChildSchema(childSchema, types, mediaType, isBodyNotJsonXml)
		schema.Items = childSchema
	case reflect.Struct:
		key := fmt.Sprintf("%v.%v", tyInfo._type.PkgPath(), tyInfo._type.Name())
		if key != "." {
			stInfo := h.handle.structs[key]
			schemaKey := stInfo.openapiName
			if h.isMullMediaType {
				schemaKey = stInfo.openapiName + "_" + mediaTypeToTypeMap[mediaType]
			}
			schema.Ref = "#/components/schemas/" + schemaKey
			return
		}
		key = fmt.Sprintf("%s%p", prefixTempStruct, tyInfo._type)
		stInfo := h.handle.structs[key]
		properties, requiredMap := h.setStructSchema(stInfo.fields)
		schema.Required = requiredMap[mediaType]
		schema.Properties = properties[mediaType]
	default:
	}
}

func (h *handlerOpenAPI) getStructBaseName(structName string) string {
	if idx := strings.Index(structName, "["); idx != -1 {
		structName = structName[:idx]
	}
	if idx := strings.LastIndex(structName, "."); idx != -1 {
		structName = structName[idx+1:]
	}
	return structName
}

func (h *handlerOpenAPI) mergeTag(tag *fieldTagInfo, fType typeInfo) {
	if fType.typeStr != "integer" {
		return
	}
	if tag.lte == nil {
		tag.lte = fType.lte
	} else if fType.lte != nil && *fType.lte < *tag.lte {
		tag.lte = fType.lte
	}
	if tag.gte == nil {
		tag.gte = fType.gte
	} else if fType.gte != nil && *fType.gte > *tag.gte {
		tag.gte = fType.gte
	}
}
