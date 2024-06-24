package goapi

import "net/http"

type APP interface {
	Init()
	GET(path string, callback func(req *http.Request, writer http.ResponseWriter))
	POST(path string, callback func(req *http.Request, writer http.ResponseWriter))
	PUT(path string, callback func(req *http.Request, writer http.ResponseWriter))
	DELETE(path string, callback func(req *http.Request, writer http.ResponseWriter))
	OPTIONS(path string, callback func(req *http.Request, writer http.ResponseWriter))
	HEAD(path string, callback func(req *http.Request, writer http.ResponseWriter))
	PATCH(path string, callback func(req *http.Request, writer http.ResponseWriter))
	TRACE(path string, callback func(req *http.Request, writer http.ResponseWriter))
	Run(addr ...string) error
}

type HandleFunc func(ctx *Context)
