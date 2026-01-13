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
	name       string // field name
	anonymous  bool
	pkgName    string
	index      int
	tag        *paramTag
	fields     []*paramField
	xmlName    string // The name of the xml structure
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
	xmlName     string // The name of the xml structure
}

type inParam struct {
	inType       InType
	parentInType InType
	values       paramFieldNames
	deeps        []int
	structField  reflect.StructField
	field        *paramField
	example      any
}

type outParam struct {
	structField reflect.StructField
	field       *paramField
	httpStatus  int
	httpHeader  http.Header
	example     any
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

type returnObjGroup struct {
	middlewares []HandleFunc
}

type returnObjDocs struct {
	info    *openapi.Info
	isDocs  bool
	tags    []*openapi.Tag
	servers []*openapi.Server
	swagger swagger.Config
}

type returnObjChild struct {
	redirectTrailingSlash  bool
	handleMethodNotAllowed bool
	noRoute                func(ctx *Context)
	noMethod               func(ctx *Context)
}

type returnObjResult struct {
	paths      []*pathInfo
	groupMap   map[string]returnObjGroup
	docsMap    map[string]returnObjDocs
	childMap   map[string]returnObjChild
	mediaTypes map[MediaType]struct{}
}

type exceptInfo struct {
	HttpCode int               `json:"http_code"`
	Header   map[string]string `json:"header"`
	Detail   string            `json:"detail"`
}
