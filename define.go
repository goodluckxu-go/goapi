package goapi

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

const (
	MethodQuery = "QUERY"
)

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
	// Other assignable parameters
	inTypeCtx InType = "Ctx" //  goapi.Context
)

const returnMediaTypeField = "media_type"

type mediaTypes struct {
	mediaTypeTagMap      map[string]MediaType
	tagMediaTypeMap      map[MediaType]string
	analysisMediaTypeMap map[MediaType]MediaTypeAnalysis
}

func (m mediaTypes) setMediaTypeAnalysis(mediaTypeAnalysis MediaTypeAnalysis) {
	mediaType, tag := mediaTypeAnalysis.Info()
	if mediaType == "" {
		return
	}
	m.tagMediaTypeMap[mediaType] = tag
	m.analysisMediaTypeMap[mediaType] = mediaTypeAnalysis
	if tag != "" {
		m.mediaTypeTagMap[tag] = mediaType
	} else if oldTag := m.tagMediaTypeMap[mediaType]; oldTag != "" {
		delete(m.mediaTypeTagMap, oldTag)
	}
}

func (m mediaTypes) getMediaTypeAnalysis(mediaType MediaType) MediaTypeAnalysis {
	return m.analysisMediaTypeMap[mediaType]
}

func (m mediaTypes) getTag(mediaType MediaType) string {
	return m.tagMediaTypeMap[mediaType]
}

func (m mediaTypes) getMediaType(tag string) MediaType {
	return m.mediaTypeTagMap[tag]
}

var allMediaType = mediaTypes{
	mediaTypeTagMap:      map[string]MediaType{},
	tagMediaTypeMap:      map[MediaType]string{},
	analysisMediaTypeMap: map[MediaType]MediaTypeAnalysis{},
}

type MediaType string

var bufPool sync.Pool

func init() {
	bufPool.New = func() any {
		return &bytes.Buffer{}
	}
}

func (m MediaType) Tag() string {
	if rs := allMediaType.getTag(m); rs != "" {
		return rs
	}
	return allMediaType.getTag(m.MediaType())
}

func (m MediaType) MediaType() MediaType {
	mediaTypeStr := string(m)
	if rs := allMediaType.getMediaType(mediaTypeStr); rs != "" {
		return rs
	}
	mediaTypeStr, _, _ = strings.Cut(mediaTypeStr, ";")
	return MediaType(mediaTypeStr)
}

func (m MediaType) DefaultName(name string) string {
	if analysis := allMediaType.getMediaTypeAnalysis(m); analysis != nil {
		return analysis.DefaultName(name)
	}
	return name
}

func (m MediaType) IsStream() bool {
	if m.Tag() == "" {
		return true
	}
	return false
}

func (m MediaType) Marshaler(v any) ([]byte, error) {
	if analysis := allMediaType.getMediaTypeAnalysis(m); analysis != nil {
		return analysis.Marshal(v)
	}
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	}
	return []byte(toString(v)), nil
}

func (m MediaType) Unmarshaler(body io.ReadCloser, value reflect.Value) error {
	if analysis := allMediaType.getMediaTypeAnalysis(m); analysis != nil {
		buf := bufPool.Get().(*bytes.Buffer)
		defer func() {
			_ = body.Close()
			buf.Reset()
			bufPool.Put(buf)
		}()
		if _, err := io.Copy(buf, body); err != nil {
			return err
		}
		return analysis.Unmarshal(buf.Bytes(), value.Interface())
	}
	if !value.IsValid() || value.Kind() != reflect.Ptr || value.IsNil() {
		_ = body.Close()
		return fmt.Errorf("MediaType: value must be a non-nil pointer")
	}
	value = value.Elem()
	if value.Type().ConvertibleTo(typeBytes) {
		defer body.Close()
		buf, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(buf).Convert(value.Type()))
		return nil
	}
	if value.Kind() == reflect.String {
		defer body.Close()
		buf, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		value.SetString(string(buf))
		return nil
	}
	if value.Type() == typeReadCloser {
		value.Set(reflect.ValueOf(body))
		return nil
	}
	_ = body.Close()
	return fmt.Errorf("MediaType: unknown type: %s", m)
}

const (
	JSON           MediaType = "application/json"
	XML            MediaType = "application/xml"
	YAML           MediaType = "application/yaml"
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
	tagName       = "name"       // Alias during verification, if not present, use desc
	tagPaths      = "paths"
	tagMethods    = "methods"
	tagSummary    = "summary"
	tagTags       = "tags"
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
var typeReadCloser = reflect.TypeOf((*io.ReadCloser)(nil)).Elem()
var typeError = reflect.TypeOf((*error)(nil)).Elem()

// inTypeCookie
var typeCookie = reflect.TypeOf(&http.Cookie{})

const (
	validErrorCode = 422
	authErrorCode  = 401
)

const defaultErrorCode = 400

var defaultNoRoute = func(ctx *Context) {
	http.Error(ctx.Writer, "404 page not found", http.StatusNotFound)
}

var defaultNoMethod = func(ctx *Context) {
	http.Error(ctx.Writer, "405 method not allowed", http.StatusMethodNotAllowed)
}

type defaultHTTPError struct {
	code int
	msg  string
}

type defaultError struct {
	XMLName xml.Name `json:"-" xml:"error" yaml:"-"`
	Error   string   `json:"error" xml:",innerxml"`
}

func (d defaultHTTPError) GetStatus() int {
	return d.code
}

func (d defaultHTTPError) GetBody() any {
	return defaultError{Error: d.msg}
}

var defaultErrorFunc = func(err error) any {
	if err == nil {
		return nil
	}
	switch val := err.(type) {
	case *HTTPError:
		return defaultHTTPError{code: val.Code, msg: val.Message}
	}
	return defaultHTTPError{code: defaultErrorCode, msg: err.Error()}
}
