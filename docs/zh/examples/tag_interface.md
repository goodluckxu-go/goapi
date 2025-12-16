## [<<](examples.md) 用接口的方式定义字段验证及文档展示
- 定义类型
- 请求和返回使用类型
- **int**表示int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64
- **float**表示float32,float64
- **all**表示所有类型
### 正则表达式
- 类型验证 **string**
- 文档可展示
~~~go
// 实现接口
type TagRegexp interface {
	Regexp() string
}

type Phone string

func (p Phone)Regexp()string  {
	return `^1{\d}10$`
}
~~~
### 枚举
- 类型验证 **string** **int** **float** **bool**
- 文档可展示
~~~go
// 实现接口
type TagEnum interface {
	Enum() []any
}

type State int

func (s State)Enum() []any  {
	return []any{1,2}
}
~~~
### 小于
- 类型验证 **int** **float**
- 文档可展示
~~~go
// 实现接口
type TagLt interface {
	Lt() float64
}

type State int

func (s State)Lt() float64  {
	return 5
}
~~~
### 小于等于
- 类型验证 **int** **float**
- 文档可展示
~~~go
// 实现接口
type TagLte interface {
	Lte() float64
}

type State int

func (s State)Lte() float64  {
	return 5
}
~~~
### 大于
- 类型验证 **int** **float**
- 文档可展示
~~~go
// 实现接口
type TagGt interface {
	Gt() float64
}

type State int

func (s State)Gt() float64  {
	return 5
}
~~~
### 大于等于
- 类型验证 **int** **float**
- 文档可展示
~~~go
// 实现接口
type TagGte interface {
	Gte() float64
}

type State int

func (s State)Gte() float64  {
	return 5
}
~~~
### 倍数
- 类型验证 **int** **float**
- 文档可展示
~~~go
// 实现接口
type TagMultiple interface {
	Multiple() float64
}

type State int

func (s State)Multiple() float64  {
	return 5
}
~~~
### 最大值
- 类型验证 **string** **object** **array**
- 文档可展示
~~~go
// 实现接口
type TagMax interface {
	Max() uint64
}

type State string

func (s State)Max() uint64  {
	return 5
}
~~~
### 最小值
- 类型验证 **string** **object** **array**
- 文档可展示
~~~go
// 实现接口
type TagMin interface {
	Min() uint64
}

type State string

func (s State)Min() uint64  {
	return 5
}
~~~
### 唯一值
- 类型验证 **[]string** **[]int** **[]float** **[]bool**
- 文档可展示
~~~go
// 实现接口
type TagUnique interface {
	Unique() bool
}

type State []string

func (s State)Unique() bool  {
	return true
}
~~~
### 描述
- 类型验证 **all**
- 文档可展示
~~~go
// 实现接口
type TagDesc interface {
	Desc() string
}

type State string

func (s State)Desc() string  {
	return "状态"
}
~~~
### 默认值
- 类型验证 类型本身
- 文档可展示
~~~go
// 实现接口
type TagDefault interface {
    Default() any
}

type State string

func (s State)Default() any  {
	return "ok"
}
~~~
### 示例
- 类型验证 类型本身
- 文档可展示
~~~go
// 实现接口
type TagExample interface {
	Example() any
}

type State string

func (s State)Example() any  {
	return "ok"
}
~~~
### 废弃字段
- 类型验证 **all**
- 文档可展示
~~~go
// 实现接口
type TagDeprecated interface {
	Deprecated() bool
}

type State string

func (s State)Deprecated() bool  {
	return true
}
~~~