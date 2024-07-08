package app

import (
	"github.com/goodluckxu-go/goapi"
	"net/http"
)

type Http struct {
	mux *http.ServeMux
}

func (h *Http) Init() {
	h.mux = http.NewServeMux()
}

func (h *Http) Handle(handler func(ctx *goapi.Context)) {
	h.mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		handler(&goapi.Context{
			Request: request,
			Writer:  writer,
		})
	})
}

func (h *Http) Run(addr string) error {
	return http.ListenAndServe(addr, h.mux)
}
