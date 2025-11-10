package goapi

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
)

type Middleware func(ctx *Context)

type InType string

func (InType) List() []InType {
	return []InType{
		inTypeHeader,
		inTypeCookie,
		inTypePath,
		inTypeQuery,
		inTypeForm,
		inTypeFile,
		inTypeBody,
	}
}

func (i InType) IsSingle() bool {
	if inArray(i, []InType{inTypePath, inTypeQuery, inTypeHeader, inTypeCookie, inTypeForm, inTypeFile}) {
		return true
	}
	return false
}

func (i InType) Tag() string {
	return string(i)
}

const (
	// single value
	inTypePath   InType = "path"
	inTypeQuery  InType = "query"
	inTypeHeader InType = "header"
	inTypeCookie InType = "cookie"
	inTypeForm   InType = "form" // default application/x-www-form-urlencoded, If a inTypeFile exists, it becomes 'multipart/form-data'
	inTypeFile   InType = "file" // default multipart/form-data
	// multiple value
	inTypeBody InType = "body"
	// security
	inTypeSecurityHTTPBearer    InType = "HTTPBearer"
	inTypeSecurityHTTPBearerJWT InType = "HTTPBearerJWT"
	inTypeSecurityHTTPBasic     InType = "HTTPBasic"
	inTypeSecurityApiKey        InType = "ApiKey"
	// Other assignable parameters, for example: goapi.Context
	inTypeOther InType = "other"
)

const returnMediaTypeField = "media_type"

type MediaType string

func (m MediaType) Tag() string {
	switch m {
	case JSON, "json":
		return "json"
	case XML, "xml":
		return "xml"
	}
	return ""
}

func (m MediaType) IsStream() bool {
	if m.Tag() == "" {
		return true
	}
	return false
}

func (m MediaType) Marshaler(v any) ([]byte, error) {
	switch m {
	case JSON:
		return json.Marshal(v)
	case XML:
		b := new(bytes.Buffer)
		b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
		body, err := xml.Marshal(v)
		if err != nil {
			return nil, err
		}
		b.Write(body)
		return b.Bytes(), nil
	}
	return []byte(toString(v)), nil
}

func (m MediaType) Unmarshaler(body io.ReadCloser, value reflect.Value) error {
	switch m {
	case JSON:
		return json.NewDecoder(body).Decode(value.Interface())
	case XML:
		return xml.NewDecoder(body).Decode(value.Interface())
	default:
		value = value.Elem()
		if value.Type().ConvertibleTo(typeBytes) {
			buf, err := io.ReadAll(body)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(buf).Convert(value.Type()))
			return nil
		}
		if value.Type().Implements(interfaceIoReadCloser) {
			value.Set(reflect.ValueOf(body).Convert(value.Type()))
			return nil
		}
	}
	return fmt.Errorf("MediaType: unknown type: %s", m)
}

const (
	JSON           MediaType = "application/json"
	XML            MediaType = "application/xml"
	formUrlencoded MediaType = "application/x-www-form-urlencoded"
	formMultipart  MediaType = "multipart/form-data"
)

const (
	tagRegexp     = "regexp"     // VALIDATION. openapi's pattern
	tagDesc       = "desc"       // openapi's description
	tagEnum       = "enum"       // openapi's enum
	tagDefault    = "default"    // openapi's default
	tagExample    = "example"    // openapi's example
	tagDeprecated = "deprecated" // openapi's deprecated
	tagLt         = "lt"         // VALIDATION. openapi's exclusiveMaximum
	tagLte        = "lte"        // VALIDATION. openapi's maximum
	tagGt         = "gt"         // VALIDATION. openapi's exclusiveMinimum
	tagGte        = "gte"        // VALIDATION. openapi's minimum
	tagMultiple   = "multiple"   // VALIDATION. openapi's multipleOf
	tagMax        = "max"        // VALIDATION. openapi's maxLength,maxItems,maxProperties
	tagMin        = "min"        // VALIDATION. openapi's minLength,minItems,minProperties
	tagUnique     = "unique"     // VALIDATION. openapi's uniqueItems
	tagPath       = "path"
	tagMethod     = "method"
	tagSummary    = "summary"
	tagTags       = "tags"
)

type LogLevel uint

var logLevel = LogInfo | LogDebug | LogWarning | LogError | LogFail

const (
	LogInfo LogLevel = 1 << iota
	LogDebug
	LogWarning
	LogError
	LogFail
)

var (
	securityTypeHTTPBearer    = reflect.TypeOf(new(HTTPBearer)).Elem()
	securityTypeHTTPBearerJWT = reflect.TypeOf(new(HTTPBearerJWT)).Elem()
	securityTypeHTTPBasic     = reflect.TypeOf(new(HTTPBasic)).Elem()
	securityTypeApiKey        = reflect.TypeOf(new(ApiKey)).Elem()
)

const omitempty = "omitempty"

var typeContext = reflect.TypeOf(&Context{})

// inTypeFile
var typeFile = reflect.TypeOf(&multipart.FileHeader{})
var typeBytes = reflect.TypeOf([]byte{})

// inTypeCookie
var typeCookie = reflect.TypeOf(&http.Cookie{})

var interfaceIoReadCloser = reflect.TypeOf(new(io.ReadCloser)).Elem()

var interfaceToStringer = reflect.TypeOf(new(fmt.Stringer)).Elem()

var typeAny = reflect.TypeOf(new(any)).Elem()

const (
	validErrorCode = 422
	authErrorCode  = 401
)
