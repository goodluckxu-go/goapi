package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

type Gin struct {
	Engine *gin.Engine
}

func (g *Gin) Init() {
	if g.Engine == nil {
		g.Engine = gin.Default()
	}
}

func (g *Gin) GET(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.Engine.GET(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) POST(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.Engine.POST(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) PUT(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.Engine.PUT(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) DELETE(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.Engine.DELETE(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) OPTIONS(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.Engine.OPTIONS(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) HEAD(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.Engine.HEAD(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) PATCH(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.Engine.PATCH(path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) TRACE(path string, callback func(req *http.Request, writer http.ResponseWriter)) {
	path = regexp.MustCompile(`\{(.*?)\}`).ReplaceAllString(path, ":$1")
	g.Engine.Match([]string{http.MethodTrace}, path, func(ctx *gin.Context) {
		callback(ctx.Request, ctx.Writer)
	})
}

func (g *Gin) Run(addr ...string) error {
	return g.Engine.Run(addr...)
}
