package goapi

import (
	"fmt"
	"net/http"
	"reflect"
)

type staticInfo struct {
	path        string
	fs          http.FileSystem
	isFile      bool
	groupPrefix string
	middlewares []HandleFunc
}

func (h *staticInfo) returnObj() (obj returnObjResult, err error) {
	h.path = pathJoin(h.groupPrefix, h.path)
	if !h.isFile {
		if h.path[len(h.path)-1] != '/' {
			h.path += "/"
		}
	}
	fsType := reflect.TypeOf(h.fs)
	pos := fmt.Sprintf("%v.%v (fs)", fsType.PkgPath(), fsType.Name())
	if fsType.Kind() == reflect.Ptr {
		fsType = fsType.Elem()
		pos = fmt.Sprintf("%v.(*%v) (fs)", fsType.PkgPath(), fsType.Name())
	}
	if len(h.middlewares) > 0 {
		pos += fmt.Sprintf(" (%v Middleware)", len(h.middlewares))
	}
	paths := []string{h.path}
	if !h.isFile {
		h.path += "{filepath:*}"
		paths = append(paths, h.path)
	}
	obj.paths = append(obj.paths, &pathInfo{
		paths:       paths,
		methods:     []string{http.MethodHead, http.MethodGet},
		pos:         pos,
		isFile:      h.isFile,
		inFs:        h.fs,
		groupPrefix: h.groupPrefix,
		isDocs:      false,
		middlewares: h.middlewares,
	})
	return
}
