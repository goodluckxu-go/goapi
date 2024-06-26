package goapi

import (
	"strings"
)

type APP interface {
	Init()
	Handle(handler func(ctx *Context))
	Run(addr string) error
}

type HandleFunc func(ctx *Context)

type AppRouter struct {
	Path    string
	Method  string
	Handler func(ctx *Context)
}

func (a *AppRouter) IsMatch(urlPath string) bool {
	pathList := strings.Split(a.Path, "/")
	relPathList := strings.Split(urlPath, "/")
	if len(pathList) != len(relPathList) {
		return false
	}
	for k, v := range pathList {
		relV := relPathList[k]
		left := strings.Index(v, "{")
		right := strings.Index(v, "}")
		if left != -1 && right != -1 {
			right = len(v) - (right + 1)
			if v[:left] != relPathList[k][:left] || v[len(v)-right:] != relPathList[k][len(relV)-right:] {
				return false
			}
		} else if relV != v {
			return false
		}
	}
	return true
}
