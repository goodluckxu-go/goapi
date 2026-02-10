## [<<](examples.md) 如何处理结构体中定义文档时无法使用复杂文本
- 使用 **{{** **}}** 来表示变量
- 定义映射时可以嵌套，但是不能嵌套成死循环
- 只能实现标签 **desc**，**summary** 替换
### 定义标签映射
~~~go
func main() {
	api := goapi.Default(true)
	api.SetStructTagVariableMapping(map[string]string{
		"prefix":   "test", 
		"username": "{{prefix}}zs", 
		"password": "{{prefix}}123456",
		"summary":  "首页文档", 
		"indexDocs": `文档说明，可以是md`,
	})
}
~~~
### 使用标签映射
~~~go
func (*Index) Index(input struct{
	router goapi.Router `paths:"/index" methods:"GET" summary:"{{summary}}" desc:"{{indexDocs}}"`
	Username string `form:"username" desc:"{{username}}"`
	Password string `form:"password" desc:"{{password}}"`
}) {

}
~~~
### 结果
1. goapi.Router行标签 **summary** 替换为 **首页文档**
2. goapi.Router行标签 **desc** 替换为 **文档说明，可以是md**
3. Username行标签 **desc** 注释替换为 **testzs**
4. Password行标签 **desc** 注释替换为 **test123456**