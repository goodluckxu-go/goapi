## [<<](examples.md) 如何使用多个程序模块
~~~go
func main() {
	api := goapi.Default(true)
	api.IsDocs = false // 默认为true，是否展示docs文档
	admin:=api.Child("/admin", "/admin")
	{
		admin.IsDocs = true // 默认为true，是否展示docs文档
		admin.OpenAPIInfo.Title = "后台管理接口"
		admin.AddMiddleware(nil)
	}
	user:=api.Child("/user", "/v1")
	{
		user.IsDocs = false
		user.OpenAPIInfo.Title = "前端小程序接口"
		user.AddMiddleware(nil)
	}
}
~~~