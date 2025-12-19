package goapi

import (
	"fmt"
	"testing"
)

func TestSetTree(t *testing.T) {
	fmt.Println("TestSetTree")
	nd := new(node)
	//nd.addRoute("/docs/", func(ctx *Context) {})
	//nd.addRoute("/docs/{path}", func(ctx *Context) {})
	//nd.addRoute("/docs/admin/", func(ctx *Context) {})
	//nd.addRoute("/docs/admin/{path}", func(ctx *Context) {})
	err := nd.addRoute("/{*}", func(ctx *Context) {})
	fmt.Println(err)
	fmt.Println(nd.getValue("/docs/admin"))
}
