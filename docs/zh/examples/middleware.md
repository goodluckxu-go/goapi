## [<<](examples.md) 如何使用中间件
### 定义中间件
~~~go
func Middleware() func(ctx *goapi.Context) {
	return func(ctx *goapi.Context) {
		// 请求前处理 
		//... 
		//ctx.Next() 
		//请求后处理 
		//... 
	}
}
~~~
### 定义运行程序
~~~go
api := goapi.Default(true)
~~~
使用中间件
~~~go
api.AddMiddleware(Middleware())
~~~
组使用中间件
~~~go
group := api.Group("/user", true)
group.AddMiddleware()
~~~
子项目使用中间件
~~~go
child := api.Child("/user", "/user")
child.AddMiddleware()
~~~

### 使用内置中间件
内置中间件位于 `github.com/goodluckxu-go/goapi/v2/middleware` 包中。

~~~go
import (
	"time"

	"github.com/goodluckxu-go/goapi/v2"
	"github.com/goodluckxu-go/goapi/v2/middleware"
)
~~~

#### 跨域中间件
使用默认跨域配置：

~~~go
api.AddMiddleware(middleware.CORSMiddleware())
~~~

自定义跨域配置：

~~~go
api.AddMiddleware(middleware.CORSMiddlewareWithConfig(middleware.CORSConfig{
	AllowOrigins:     []string{"https://example.com"},
	AllowMethods:     []string{"GET", "POST", "OPTIONS"},
	AllowHeaders:     []string{"Content-Type", "Authorization"},
	ExposeHeaders:    []string{"X-Request-ID"},
	AllowCredentials: true,
	MaxAge:           12 * time.Hour,
}))
~~~

`AllowCredentials` 为 `true` 时，需要配置明确的 `AllowOrigins` 或 `AllowOriginFunc`，不能使用通配符 `*`。

#### 限流中间件
限制同一个客户端在指定时间窗口内的请求次数：

~~~go
api.AddMiddleware(middleware.RateLimitMiddleware(100, time.Minute))
~~~

自定义限流配置：

~~~go
api.AddMiddleware(middleware.RateLimitMiddlewareWithConfig(middleware.RateLimitConfig{
	Limit:  100,
	Window: time.Minute,
	Burst:  20,
	KeyFunc: func(ctx *goapi.Context) string {
		return ctx.RemoteIP()
	},
	Message:         "too many requests",
	CleanupInterval: 5 * time.Minute,
	MaxKeys:         10000,
}))
~~~

`Limit` 和 `Window` 表示平均限流速度，`Burst` 表示短时间内允许的最大突发请求数。默认按客户端直连 IP 限流，也可以通过 `KeyFunc` 改成按用户 ID、Token 或接口路径等维度限流。超过限制时返回 `429 Too Many Requests`。

#### 请求体大小限制中间件
限制请求体最大字节数：

~~~go
api.AddMiddleware(middleware.BodyLimitMiddleware(10 << 20)) // 10MB
~~~

自定义请求体限制配置：

~~~go
api.AddMiddleware(middleware.BodyLimitMiddlewareWithConfig(middleware.BodyLimitConfig{
	Limit:   10 << 20,
	Message: "request body too large",
}))
~~~

`Limit` 的单位是字节；当 `Limit` 为 `0` 时，只允许空请求体。超过限制时返回 `413 Request Entity Too Large`。
