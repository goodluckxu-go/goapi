## [<<](examples.md) 用结构体标签的方式定义字段验证及文档展示
- **int**表示int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64
- **float**表示float32,float64
- **all**表示所有类型
### 正则表达式
- 类型验证 **string**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" regexp:"^id_\d+$"`
}
~~~
### 枚举
- 以,分割列表
- 类型验证 **string** **int** **float** **bool**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" enum:"zhangsan,lisi"`
}
~~~
### 小于
- 类型验证 **int** **float**
- 文档可展示
~~~go
type Body struct {
	Name int `json:"name" lt:"10"`
}
~~~
### 小于等于
- 类型验证 **int** **float**
- 文档可展示
~~~go
type Body struct {
	Name int `json:"name" lte:"10"`
}
~~~
### 大于
- 类型验证 **int** **float**
- 文档可展示
~~~go
type Body struct {
	Name int `json:"name" gt:"10"`
}
~~~
### 大于等于
- 类型验证 **int** **float**
- 文档可展示
~~~go
type Body struct {
	Name int `json:"name" gte:"10"`
}
~~~
### 倍数
- 类型验证 **int** **float**
- 文档可展示
~~~go
type Body struct {
	Name int `json:"name" multiple:"10"` // name只能为10*n
}
~~~
### 最大值
- 类型验证 **string** **object** **array**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" max:"10"`
}
~~~
### 最小值
- 类型验证 **string** **object** **array**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" min:"15"`
}
~~~
### 唯一值
- 类型验证 **[]string** **[]int** **[]float** **[]bool**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" unique:"true"`
}
~~~
### 描述
- 类型验证 **all**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" desc:"名称"`
}
~~~
### 默认值
- 类型验证 **string** **int** **float** **bool**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" default:"zhangsan"`
}
~~~
### 示例
- 类型验证 **string** **int** **float** **bool**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" example:"zhangsan"`
}
~~~
### 废弃字段
- 类型验证 **all**
- 文档可展示
~~~go
type Body struct {
	Name string `json:"name" deprecated:"true"`
}
~~~