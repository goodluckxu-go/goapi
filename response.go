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
	Status() int
}

type ResponseStatusText interface {
	StatusText() string
}

type ResponseHeader interface {
	Header() http.Header
}

type ResponseBody interface {
	Body() any
}
