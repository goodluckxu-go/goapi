package app

import (
	"github.com/gin-gonic/gin"
	"github.com/goodluckxu-go/goapi"
)

type Gin struct {
	engine *gin.Engine
}

func (g *Gin) Init() {
	if g.engine == nil {
		gin.SetMode(gin.ReleaseMode)
		g.engine = gin.New()
	}
}

func (g *Gin) Handle(handler func(ctx *goapi.Context)) {
	g.engine.Any("/*path", func(ctx *gin.Context) {
		handler(&goapi.Context{
			Request: ctx.Request,
			Writer:  ctx.Writer,
		})
	})
}

func (g *Gin) Run(addr string) error {
	return g.engine.Run(addr)
}
