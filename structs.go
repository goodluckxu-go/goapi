package goapi

import (
	"net/http"
	"reflect"

	"github.com/goodluckxu-go/goapi/openapi"
	"github.com/goodluckxu-go/goapi/swagger"
)

type paramTag struct {
	desc       string
	_default   any
	example    any
	deprecated bool
	// valid
	regexp   string
	enum     []any
	lt       *float64
	lte      *float64
	gt       *float64
	gte      *float64
	multiple *float64
	max      *uint64
	min      uint64
	unique   bool
}

type paramField struct {
	_type      reflect.Type
	kind       reflect.Kind
	isTextType bool // Whether to inherit TextInterface
	names      paramFieldNames
	anonymous  bool
	pkgName    string
	index      int
	tag        *paramTag
	fields     []*paramField
}

type paramFieldName struct {
	required  bool
	name      string
	split     paramFieldNameSplit
	mediaType MediaType
	inType    InType
}

type paramFieldNameSplit []string

type paramFieldNames []paramFieldName

func (p paramFieldNames) getFieldName(mediaType MediaType) (fieldName paramFieldName) {
	for _, n := range p {
		if n.mediaType == mediaType {
			return n
		}
	}
	return
}

func (p paramFieldNames) MediaTypes() (rs []MediaType) {
	for _, n := range p {
		rs = append(rs, n.mediaType)
	}
	return
}

type structInfo struct {
	_type       reflect.Type
	openapiName string
	fields      []*paramField
}

type inParam struct {
	inType       InType
	parentInType InType
	values       paramFieldNames
	deeps        []int
	structField  reflect.StructField
	field        *paramField
}

type outParam struct {
	structField reflect.StructField
	field       *paramField
	httpStatus  int
	httpHeader  http.Header
}

type pathInfo struct {
	paths   []string
	methods []string
	pos     string
	handle  HandleFunc
	// call
	value       reflect.Value
	inFs        http.FileSystem // file
	isFile      bool            // file
	inTypes     []reflect.Type  // func in types
	inParams    []*inParam
	outParam    *outParam
	middlewares []HandleFunc
	// openapi
	summary     string
	desc        string
	tags        []string
	docsPath    string
	isDocs      bool
	groupPrefix string
	isSwagger   bool
}

type pathInterfaceResult struct {
	paths             []*pathInfo
	publicMiddlewares map[string][]HandleFunc
	mediaTypes        map[MediaType]struct{}
	openapiMap        map[string]*openapi.OpenAPI
	tags              []*openapi.Tag
}

type pathInterface interface {
	// input is parent params
	// docsPath is ChildAPI input
	// groupPrefix is ChildAPI or APIGroup input
	returnObj(prefix, docsPath, groupPrefix string, middlewares []HandleFunc, isDocs bool) (obj pathInterfaceResult, err error)
}

type returnObjGroup struct {
	middlewares []HandleFunc
}

type returnObjDocs struct {
	info    *openapi.Info
	tags    []*openapi.Tag
	servers []*openapi.Server
	swagger swagger.Config
}

type returnObjResult struct {
	paths      []*pathInfo
	groupMap   map[string]returnObjGroup
	docsMap    map[string]returnObjDocs
	mediaTypes map[MediaType]struct{}
}

type exceptInfo struct {
	HttpCode int               `json:"http_code"`
	Header   map[string]string `json:"header"`
	Detail   string            `json:"detail"`
}
