package goapi

type APP interface {
	Init()
	Handle(handler func(ctx *Context))
	Run(addr string) error
}
