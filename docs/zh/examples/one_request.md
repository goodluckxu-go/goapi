## [<<](examples.md) 如何定义一个请求
### 定义一个结构体
~~~go
type Index struct {}
~~~
### 给结构体定义一个方法
- **paths** 访问路径，可以多个，以 **,** 分割，如：**paths:"/index,/page"** 表示请求路径可以是 **/index** 和 **/page**
- **methods** 请求方法，可以多个，以 **,** 分割，如 **methods:"GET,POST"** 表示可以以 **GET** 和 **POST** 请求
- **summary** 概要，在 **swagger** 路由同行展示
- **desc** 描述，在 **swagger** 路由展开后展示
- **tags** 标签，可以多个，以 **,** 分割，，用于 **swagger** 标签分组，详见：[文档定义标签分组](docs_tags.md)
~~~go
// 定义一个类型为 goapi.Router 的字段
// 定义必要标签 paths 和 methods
func (*Index) Index(input struct{
	router goapi.Router `paths:"/index,/page" methods:"GET,POST" summary:"测试" desc:"测试" tags:"user,admin"`
}) {

}
~~~
### 结构体方法中使用上下文
~~~go
// 上下文必须在匿名结构体前面
// 定义一个类型为 goapi.Router 的字段
// 定义必要标签 paths 和 methods
func (*Index) Index(ctx *goapi.Context, input struct{
	router goapi.Router `paths:"/index,/page" methods:"GET,POST" summary:"测试" desc:"测试" tags:"user,admin"`
}) {

}
~~~
### 引入结构体
~~~go
api := goapi.GoAPI(true)
api.IncludeRouter(&PersonController{}, "/v1", true)
~~~