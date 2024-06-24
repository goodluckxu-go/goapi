package goapi

import "reflect"

type includeRouter struct {
	router      any
	prefix      string
	isDocs      bool
	middlewares []Middleware
}

type routerInfo struct {
	path    string
	methods []string
	summary string
	desc    string
	tags    []string
}

type pathInfo struct {
	path           string
	methods        []string
	inTypes        []reflect.Type
	inputFields    []fieldInfo // The fields of the input struct
	middlewares    []Middleware
	respMediaTypes []MediaType
	funcValue      reflect.Value // The value of routing function
	// openapi
	summary   string
	desc      string
	tags      []string
	res       *fieldInfo // response
	exceptRes *fieldInfo // exception response
}

type structInfo struct {
	name        string
	pkg         string
	_type       reflect.Type
	fields      []fieldInfo
	openapiName string
}

type fieldInfo struct {
	name           string
	_type          reflect.Type
	inType         string
	inTypeSecurity string
	inTypeVal      string
	tag            *fieldTagInfo
	deepIdx        []int       // Struct depth index
	deepTypes      []typeInfo  // Struct type depth
	_struct        *structInfo // Exists when type is struct
	mediaTypes     []mediaTypeInfo
}

type fieldTagInfo struct {
	desc     string
	_default any
	example  any
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

type mediaTypeInfo struct {
	_type    string // media type, example 'json', 'xml'
	name     string
	required bool
}

type typeInfo struct {
	_type    reflect.Type
	typeStr  string // openapi use
	format   string // openapi use
	isStruct bool
}

type exceptInfo struct {
	HttpCode int               `json:"http_code"`
	Header   map[string]string `json:"header"`
	Detail   string            `json:"detail"`
}
