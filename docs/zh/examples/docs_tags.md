## [<<](examples.md) 文档定义标签分组
### 直接在方法中定义tags标签
- 多个标签用,分割
~~~
type Index struct{}
~~~
定义方法标签
~~~go
func (*index)List (input struct {
	router goapi.Router `paths:"/list" methods:"GET" tags:"user,admin"`
}) {
	
}
~~~
给标签添加注释
~~~go
api := goapi.GoAPI(true)
api.OpenAPITags = []*openapi.Tag{
	{Name: "user", Description: "用户组"},
	{Name: "admin", Description: "管理员组"},
}
~~~
### 用接口的方式
所有在Index结构体中的方法都使用
~~~go
func (*Index)Tags()[]*openapi.Tag {
	return []*openapi.Tag{
		{Name: "user", Description: "用户组"},
		{Name: "admin", Description: "管理员组"},
	}
}
~~~