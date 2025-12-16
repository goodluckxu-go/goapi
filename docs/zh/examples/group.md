## [<<](examples.md) 如何使用组
### 程序定义组
- 文档访问路径 /docs/
- 路由前缀路径  /admin
~~~go
func main () {
	api := goapi.GoAPI(true)
	admin:=api.Group("/admin", true)
	{
		admin.AddMiddleware(nil)
	}
}
~~~
### 子程序定义组
- 文档访问路径 /docs/admin/
- 路由前缀路径  /v1/user
~~~go
func main () {
	api := goapi.GoAPI(true)
	admin:=api.Child("/admin", true, "/v1")
	{
		admin.OpenAPIInfo.Title = "后台管理接口"
		admin.AddMiddleware(nil)
		group:=admin.Group("/user", true)
		{
			group.AddMiddleware(nil)
		}
	}
}
~~~
### 组下面定义组
- 文档访问路径 /docs/
- 路由前缀路径  /admin/user
~~~go
func main () {
	api := goapi.GoAPI(true)
	admin:=api.Group("/admin", true)
	{
		admin.AddMiddleware(nil)
		user:=admin.Group("/user", true)
		{
			user.AddMiddleware(nil)
		}
	}
}
~~~