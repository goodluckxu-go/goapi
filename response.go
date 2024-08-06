package goapi

import (
	"net/http"
)

type Response interface {
	GetBody() any
	GetContentType() string
	SetContentType(contentType string)
	Write(w http.ResponseWriter)
}
