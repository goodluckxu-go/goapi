package app

import (
	"fmt"
	"net/http"
	"strings"
)

type Http struct {
	routers []httpRouter
}

func (h *Http) Init() {
}

func (h *Http) GET(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	h.routers = append(h.routers, httpRouter{path: path, callback: callback})
}

func (h *Http) POST(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	h.routers = append(h.routers, httpRouter{path: path, callback: callback})
}

func (h *Http) PUT(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	h.routers = append(h.routers, httpRouter{path: path, callback: callback})
}

func (h *Http) DELETE(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	h.routers = append(h.routers, httpRouter{path: path, callback: callback})
}

func (h *Http) OPTIONS(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	h.routers = append(h.routers, httpRouter{path: path, callback: callback})
}

func (h *Http) HEAD(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	h.routers = append(h.routers, httpRouter{path: path, callback: callback})
}

func (h *Http) PATCH(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	h.routers = append(h.routers, httpRouter{path: path, callback: callback})
}

func (h *Http) TRACE(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	h.routers = append(h.routers, httpRouter{path: path, callback: callback})
}

func (h *Http) Run(addr ...string) error {
	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		count := 0
		for _, router := range h.routers {
			_, err := h.getPaths(router.path, req.URL.Path)
			if err != nil {
				continue
			}
			if count > 0 {
				panic("path repeat")
			}
			router.callback(req, writer)
			count++
		}
	})
	ad := ":8080"
	if len(addr) > 0 {
		ad = addr[0]
	}
	return http.ListenAndServe(ad, nil)
}

func (h *Http) getPaths(path, urlPath string) (rs map[string]string, err error) {
	rs = map[string]string{}
	pathList := strings.Split(path, "/")
	relPathList := strings.Split(urlPath, "/")
	if len(pathList) != len(relPathList) {
		err = fmt.Errorf("path format error")
		return
	}
	for k, v := range pathList {
		relV := relPathList[k]
		left := strings.Index(v, "{")
		right := strings.Index(v, "}")
		if left != -1 && right != -1 {
			right = len(v) - (right + 1)
			if v[:left] != relPathList[k][:left] || v[len(v)-right:] != relPathList[k][len(relV)-right:] {
				err = fmt.Errorf("path format error")
				rs = nil
				return
			}
			rs[v[left+1:len(v)-right-1]] = relPathList[k][left : len(relV)-right]
		} else if relV != v {
			err = fmt.Errorf("path format error")
			rs = nil
			return
		}
	}
	return
}

type httpRouter struct {
	path     string
	callback func(req *http.Request, writer http.ResponseWriter)
}
