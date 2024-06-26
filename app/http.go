package app

import (
	"github.com/goodluckxu-go/goapi"
	"net/http"
)

type Http struct {
}

func (h *Http) Init() {
}

func (h *Http) Handle(handler func(ctx *goapi.Context)) {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		handler(&goapi.Context{
			Request: request,
			Writer:  writer,
		})
	})
}

func (h *Http) Run(addr ...string) error {
	ad := ":8080"
	if len(addr) > 0 {
		ad = addr[0]
	}
	return http.ListenAndServe(ad, nil)
}
