package goapi

import (
	"reflect"
)

type includeRouter struct {
	router      any
	prefix      string
	isDocs      bool
	middlewares []Middleware
}

type staticInfo struct {
	path string
	root string
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
	docsPath       string
	// openapi
	summary   string
	desc      string
	tags      []string
	res       *fieldInfo // response
	exceptRes *fieldInfo // exception response
	isDocs    bool
	pos       string // position
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
	mediaTypes     []MediaType
	required       bool
	fieldMap       map[MediaType]*fieldNameInfo
}

type fieldTagInfo struct {
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

type fieldNameInfo struct {
	name     string
	required bool
	xml      *xmlInfo
}

type xmlInfo struct {
	attr     bool
	innerxml bool
	childs   []string
}

type typeInfo struct {
	_type    reflect.Type
	typeStr  string // openapi use
	format   string // openapi use
	isStruct bool
	// 'int8', 'int16', 'uint8', 'uint16' verification use
	lte *float64
	gte *float64
}

type exceptInfo struct {
	HttpCode int               `json:"http_code"`
	Header   map[string]string `json:"header"`
	Detail   string            `json:"detail"`
}

type appRouter struct {
	path     string
	isPrefix bool
	method   string
	handler  func(ctx *Context)
	pos      string
}
