package goapi

import (
	"net/http"
)

type ResponseStatus interface {
	GetStatus() int
}

type ResponseHeader interface {
	GetHeader() http.Header
}

type ResponseBody interface {
	GetBody() any
}

// NewHTTPError create HTTP error
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}
