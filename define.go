package goapi

import (
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
)

const inTypeHeader = "header"
const inTypeCookie = "cookie"
const inTypePath = "path"
const inTypeQuery = "query"
const inTypeBody = "body"
const inTypeForm = "form" // default application/x-www-form-urlencoded, If a inTypeFile exists, it becomes 'multipart/form-data'
const inTypeFile = "file" // default multipart/form-data

const inTypeSecurityHTTPBearer = "HTTPBearer"
const inTypeSecurityHTTPBasic = "HTTPBasic"
const inTypeSecurityApiKey = "ApiKey"

var inTypes = []string{
	inTypeHeader,
	inTypeCookie,
	inTypePath,
	inTypeQuery,
	inTypeBody,
	inTypeForm,
	inTypeFile,
}

type MediaType string

const jsonType = "json"
const xmlType = "xml"

const omitempty = "omitempty"

const JSON MediaType = "application/json"
const XML MediaType = "application/xml"
const formUrlencoded MediaType = "application/x-www-form-urlencoded"
const formMultipart MediaType = "multipart/form-data"

var mediaTypeToTypeMap = map[MediaType]string{
	JSON: jsonType,
	XML:  xmlType,
}

var typeToMediaTypeMap = map[string]MediaType{
	jsonType: JSON,
	xmlType:  XML,
}

var bodyMediaTypes = []MediaType{
	JSON,
	XML,
}

const prefixTempStruct = "tmp_"

const tagRegexp = "regexp"     // VALIDATION. openapi's pattern
const tagDesc = "desc"         // openapi's description
const tagEnum = "enum"         // openapi's enum
const tagDefault = "default"   // openapi's default
const tagExample = "example"   // openapi's example
const tagLt = "lt"             // VALIDATION. openapi's exclusiveMaximum
const tagLte = "lte"           // VALIDATION. openapi's maximum
const tagGt = "gt"             // VALIDATION. openapi's exclusiveMinimum
const tagGte = "gte"           // VALIDATION. openapi's minimum
const tagMultiple = "multiple" // VALIDATION. openapi's multipleOf
const tagMax = "max"           // VALIDATION. openapi's maxLength,maxItems,maxProperties
const tagMin = "min"           // VALIDATION. openapi's minLength,minItems,minProperties
const tagUnique = "unique"     // VALIDATION. openapi's uniqueItems

const validErrorCode = 422

type LogLevel uint

var logLevel LogLevel = LogInfo | LogDebug | LogWarning | LogError | LogFail

const (
	LogInfo LogLevel = 1 << iota
	LogDebug
	LogWarning
	LogError
	LogFail
)

var securityTypeHTTPBearer = reflect.TypeOf(new(HTTPBearer)).Elem()
var securityTypeHTTPBasic = reflect.TypeOf(new(HTTPBasic)).Elem()
var securityTypeApiKey = reflect.TypeOf(new(ApiKey)).Elem()

var typeResponse = reflect.TypeOf(new(Response)).Elem()

var securityTypes = []reflect.Type{
	securityTypeHTTPBearer,
	securityTypeHTTPBasic,
	securityTypeApiKey,
}

var typeFile = reflect.TypeOf(&multipart.FileHeader{})
var typeFiles = reflect.TypeOf([]*multipart.FileHeader{})
var typeCookie = reflect.TypeOf(&http.Cookie{})

var typeContext = reflect.TypeOf(&Context{})

var typeBytes = reflect.TypeOf([]byte{})
var interfaceIoReadCloser = reflect.TypeOf(new(io.ReadCloser)).Elem()

var typeAny = reflect.TypeOf(new(any)).Elem()
