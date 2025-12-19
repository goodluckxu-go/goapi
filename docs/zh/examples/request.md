## [<<](examples.md) 如何定义请求值
### 使用上下文获取参数
~~~go
func (*Index) Param(ctx *goapi.Context, input struct {
	input  goapi.Router `path:"/param" method:"POST" summary:"参数请求"`
}) {
	// 获取*http.Request参数 
	request := ctx.Request 
	// 根据系统request参数获取
}
~~~
### 使用header,cookie,query,path,form,file标签定义简单字段
定义path请求
~~~go
// path参数请求，遇到/则结束
// 例如：
//  /param1/test           valid        path=test
//  /param1/test/          noValid
//  /param1/test/request   noValid
func (*Index) Param1(input struct {
	input  goapi.Router `path:"/param1/{path}" method:"POST" summary:"参数请求"`
	Path   string       `path:"path" desc:"主键定义，必填"` // path不能定义omitempty为非必填
}) {

}

// path参数请求，匹配后面所有
// 例如：
//  /param1/test           valid        path=test
//  /param1/test/          valid        path=test/
//  /param1/test/request   valid        path=test/request
func (*Index) Param2(input struct {
	input  goapi.Router `path:"/param2/{path:*}" method:"POST" summary:"参数请求"`
	Path   string       `path:"path" desc:"主键定义，必填"` // path不能定义omitempty为非必填
}) {

}
~~~
定义header,cookie,query请求
~~~go
func (*Index) Param(input struct {
	input  goapi.Router `path:"/param" method:"POST" summary:"参数请求"`
	Token1 string       `cookie:"Token1" desc:"cookie中Token1定义，必填"`
	Token2 *http.Cookie `cookie:"token2,omitempty" desc:"cookie中token2定义，非必填"`
	Token3 int64        `header:"Token3" desc:"header的Token3定义，必填"`
	Page   int          `query:"page,omitempty" desc:"query中的page定义，非必填"`
}) {

}
~~~
定义application/x-www-form-urlencoded请求，**可以form和file同时定义，请求类型会变为multipart/form-data**
~~~go
func (*Index) Form(input struct {
	input    goapi.Router `path:"/form" method:"POST" summary:"form请求"`
	Username string       `form:"username" desc:"用户名，必填"`
	Password string       `form:"password,omitempty" desc:"密码，非必填"`
}) {

}
~~~
定义multipart/form-data请求
~~~go
func (*Index) File(input struct {
	input goapi.Router            `path:"/file" method:"POST" summary:"file请求"`
	File  *multipart.FileHeader   `form:"file" desc:"文件"`
	Files []*multipart.FileHeader `file:"files" desc:"文件列表"`
}) {

}
~~~
### 使用body标签定义一个复杂的值
- body标签中定义**Content-Type**
- body标签中定义多个**Content-Type**用,分割
- application/json可以简写为json，application/xml可以简写为xml
定义application/json和application/xml请求
~~~go
type BodyReq struct {
	ID   int    `json:"ID" xml:"id" desc:"主键，必填"`     // 必须传一个不等于0的值，json传ID,xml传id
	Age  *int   `json:"Age" desc:"年龄，必填"`             // 必须传该字段，字段可以为0
	Name string `json:"Name,omitempty" desc:"名称，非必填"` // 必须传不为空字符串的值 
	Desc string `desc:"详情，必填"`                        // json和xml都传Desc字段
}

func (*Index) Post(input struct {
	input goapi.Router `path:"/post" summary:"请求"`
	Body  BodyReq      `body:"json,xml" desc:"body信息"`
}) {

}
~~~
定义其他类型请求，body里面为**Content-Type**类型
~~~go
func (*Index) PostIoReader(input struct {
	input goapi.Router     `path:"/post/io" method:"POST" summary:"请求"`
	Body  io.ReadCloser    `body:"text/html" desc:"body信息，接受一个可读取类型"`
}) {

}
func (*Index) PostByte(input struct {
	input goapi.Router     `path:"/post/byte" method:"POST" summary:"请求"`
	Body  []byte           `body:"application/octet-stream" desc:"body信息，接受一个[]byte值"`
}) {

}
func (*Index) PostString(input struct {
	input goapi.Router     `path:"/post/string" method:"POST" summary:"请求"`
	Body  string           `body:"text/plain" desc:"body信息，接受一个string类型值"`
}) {

}
~~~