package goapi

import (
	"net/http"
)

type Response interface {
	ResponseStatusCode
	ResponseHeader
	ResponseBody
}

type ResponseStatusCode interface {
	GetStatusCode() int
}

type ResponseHeader interface {
	GetHeader() http.Header
}

type ResponseBody interface {
	GetBody() any
}
