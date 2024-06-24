package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

type Gin struct {
	engine *gin.Engine
}

func (g *Gin) Init() {
	g.engine = gin.Default()
}

func (g *Gin) GET(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.engine.GET(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) POST(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.engine.POST(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) PUT(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.engine.PUT(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) DELETE(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.engine.DELETE(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) OPTIONS(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.engine.OPTIONS(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) HEAD(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.engine.HEAD(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) PATCH(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.engine.PATCH(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) TRACE(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.engine.Match([]string{http.MethodTrace}, path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) Run(addr ...string) error {
	return g.engine.Run(addr...)
}
