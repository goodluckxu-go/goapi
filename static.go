package goapi

import (
	"fmt"
	"net/http"
	"reflect"
)

type staticInfo struct {
	path   string
	fs     http.FileSystem
	isFile bool
}

func (h *staticInfo) returnObj(prefix, docsPath, groupPrefix string, middlewares []Middleware, isDocs bool) (obj pathInterfaceResult, err error) {
	h.path = pathJoin(prefix, h.path)
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
	if len(middlewares) > 0 {
		pos += fmt.Sprintf(" (%v Middleware)", len(middlewares))
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
		groupPrefix: groupPrefix,
		docsPath:    docsPath,
		isDocs:      false,
		middlewares: middlewares,
	})
	return
}
