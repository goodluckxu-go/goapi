## [<<](examples.md) 如何定义返回值
### 系统提供返回
http返回
~~~go
func (*Index) Param(input struct {
	input  goapi.Router `path:"/param" method:"POST" summary:"参数请求"`
}) *string {
	return nil
}
~~~
文件返回
~~~go
func (*Index) Param(input struct {
	input  goapi.Router `path:"/param" method:"POST" summary:"参数请求"`
}) *string {
	return nil
}
~~~
SSE返回
~~~go
func (*Index) Param(input struct {
	input  goapi.Router `path:"/param" method:"POST" summary:"参数请求"`
}) *response.SSEResponse {
	return &response.SSEResponse{
		SSEWriter: func(s *response.SSEvent) {
			for {
				s.Write(response.SSEventData{
					Event: "message", 
					Data:  "测试数据", 
					Id:    "", 
					Retry: 0,
				})
				time.Sleep(1 * time.Second)
			}
		},
	}
}
~~~
### 最基本的返回值
~~~go
type BodyResp struct {
	ID   int    `json:"ID" xml:"id" desc:"主键，必填"`     // 必须传一个不等于0的值，json传ID,xml传id
	Age  *int   `json:"Age" desc:"年龄，必填"`             // 必须传该字段，字段可以为0
	Name string `json:"Name,omitempty" desc:"名称，非必填"` // 必须传不为空字符串的值 
	Desc string `desc:"详情，必填"`                        // json和xml都传Desc字段
}

func (*Index) Post(input struct {
	input goapi.Router `path:"/post" summary:"请求"`
}) *BodyResp{
	return nil
}
~~~
### 重新定义返回值的情况
~~~go
// 实现接口可重新定义http_code返回状态，不实现默认返回200
type ResponseStatus interface {
	GetStatus() int
}

// 实现接口可重新定义header返回请求
type ResponseHeader interface {
	GetHeader() http.Header
}

// 实现接口可重新定义返回数据结构
type ResponseBody interface {
	GetBody() any
}

func (b BodyResp)GetStatus() int {
    return 201
}

func (b BodyResp)GetHeader() http.Header {
    header:=new(http.Header)
	token := &http.Cookie{
		Name:  "Token", 
		Value: "123456",
	}
	header.Add("Set-Cookie", token.String())
	header.Add("Content-Type","text/html")
	return header
}

type NewBodyResp struct (
    Code int
	Msg string
	Data BodyResp
)


func (b BodyResp)GetBody() any  {
	code:=0
	msg:=""
	var bb BodyResp
	if b.ID == 0 {
		code=400
		msg="数据不为空"
	} else {
		bb = b
	}
	return NewBodyResp{
		Code: code,
		Msg: msg,
		Data: bb,
	}
}
~~~
### 需要使用流式返回数据
- 返回值需要实现 **io.ReadCloser** 接口
- **GetBody() any** 返回一个实现 **io.ReadCloser** 接口的返回值
- 处理后会自动调用 **Close** 关闭
- 如果直接返回 **io.ReadCloser** 没有设置 **Content-Type** ，默认 **Content-Type** 为 **application/octet-stream**

基本返回值实现接口
~~~go
type StreamResp struct {
	R io.ReadCloser
}

func (s StreamResp) Read(buf []byte) (int, error) {
	return s.R.Read(buf)
}

func (s StreamResp) Close() error {
	return s.R.Close()
}
~~~
重新定义实现 **GetBody() any** 方法实现接口
~~~go
type StreamResp struct {
	R io.ReadCloser
}

func (s StreamResp) GetBody() any {
	return s.R
}
~~~