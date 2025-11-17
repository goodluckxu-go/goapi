package goapi

import (
	"net/http"
)

type Response interface {
	ResponseStatus
	ResponseHeader
	ResponseBody
}

type ResponseStatus interface {
	HttpStatus() int
}

type ResponseHeader interface {
	HttpHeader() http.Header
}

type ResponseBody interface {
	HttpBody() any
}
