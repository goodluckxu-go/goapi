package main

import (
	"github.com/goodluckxu-go/goapi"
)

func main() {
	api := goapi.GoAPI(true)
	api.AddMiddleware(func(ctx *goapi.Context) {
		// todo
		ctx.Next()
		// todo
	})
	api.SetResponseMediaType(goapi.JSON)
	api.HTTPExceptionHandler(func(httpCode int, detail string) goapi.Response {
		return &goapi.HTTPResponse[any]{
			HttpCode: httpCode,
			Body:     detail,
		}
	})
	api.IncludeRouter(&TestController{}, "/v1", true)
	api.Run("127.0.0.1:8080")
}

type TestController struct {
}

func (t *TestController) OneParam(input struct {
	router goapi.Router `path:"/oneParam" method:"put" summary:"one param router" desc:"example one param router"`
	Auth   *AdminAuth
	Body   *Req `body:"json"`
}) *Resp {
	return &Resp{
		Code: 100,
		Data: input.Body,
	}
}

func (t *TestController) TwoParam(ctx *goapi.Context, input struct {
	router    goapi.Router `path:"/twiParam/{id}" method:"get" summary:"one param router" description:"example one param router"`
	Auth      *AdminAuth
	MediaType string `query:"media_type,omitempty"`
	ID        string `path:"id"`
}) *Resp {
	ctx.Writer.WriteHeader(200)
	return &Resp{
		Code: 100,
		Data: input.ID,
	}
}

type Req struct {
	Name string `json:"name,omitempty"`
	Age  int    `json:"age"`
}

type Resp struct {
	Code int `json:"code"`
	Data any `json:"data,omitempty"`
}

type AdminAuth struct {
}

func (a *AdminAuth) HTTPBearer(token string) {
	if token != "123456" {
		goapi.HTTPException(403, "token is 123456")
	}
}
