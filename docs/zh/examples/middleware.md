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
api := goapi.GoAPI(true)
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
child := api.Child("/user", true, "/user")
child.AddMiddleware()
~~~