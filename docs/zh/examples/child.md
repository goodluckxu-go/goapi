## [<<](examples.md) 如何使用多个程序模块
~~~go
api := goapi.GoAPI(true)
admin:=api.Child("/admin", true, "/admin")
{
	admin.OpenAPIInfo.Title = "后台管理接口"
	admin.AddMiddleware(nil)
}
user:=api.Child("/user", true, "/v1")
{
	user.OpenAPIInfo.Title = "前端小程序接口"
	user.AddMiddleware(nil)
}
~~~