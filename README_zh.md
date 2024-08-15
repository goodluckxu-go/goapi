# goapi
使用OpenAPI3.1文档的HTTP框架

[English](README.md) | 中文

## 用法
~~~bash
go get github.com/goodluckxu-go/goapi
~~~
main.go
~~~go
import (
	"github.com/fatih/color"
	"github.com/goodluckxu-go/goapi"
)

func main() {
	color.NoColor = true // 关闭控制台颜色，默认颜色
	api := goapi.GoAPI(true, "/docs")
	api.SetResponseMediaType(goapi.JSON)
	api.HTTPExceptionHandler(func(httpCode int, detail string) goapi.Response {
		return &goapi.HTTPResponse[Error]{
			HttpCode: httpCode, 
			Body: Error{
				Code:  httpCode, 
				Error: detail,
			},
		}
	})
	api.IncludeRouter(&IndexController{}, "/v1", true, func(ctx *goapi.Context) {
		ctx.Next()
	})
	_ = api.Run("127.0.0.1:8080")
}
~~~
user_controller.go
~~~go
type UserController struct {
}

type UserListRouter struct {
}

type Req struct {
	Page int `json:"page" desc:"now page" gt:"0"`
	Limit int `json:"page" desc:"limit a page count" gte:"10" lte:"100"`
}

func (u *UserListRouter) Index(input struct {
	router  goapi.Router `path:"/index/{id}" method:"put" summary:"test api" desc:"test api" tags:"admin"`
	Auth    *AdminAuth
	ID      string `path:"id" regexp:"^\d+$"` // path 
	Req     Req `body:"json"`
}) Resp {
	return Resp{}
}

// 实现HTTPBearer接口
type AdminAuth struct {
	Admin  string          // 定义一个值并从控制器检索它
}

func (h *AdminAuth) HTTPBearer(token string) {
	if token != "123456" {
		response.HTTPException(401, "token is error")   
	}
	h.Admin = "admin"
}

// 实现HTTPBasic接口
type AdminAuth struct {
	Admin  string          // 定义一个值并从控制器检索它
}

func (h *AdminAuth) HTTPBasic(username,password string) {
	if username != "admin" || password != "123456" {
		response.HTTPException(401, "token is error")
	} 
	h.Admin = "admin"
}


// 实现ApiKey接口
type AdminAuth struct {
	Token  string   `header:"Token"`
	Admin  string          // 定义一个值并从控制器检索它
}

func (h *AdminAuth) ApiKey() {
	if h.Token != "123456" {
		response.HTTPException(401, "token is error")
	}
	h.Admin = "admin"
}
~~~
### 'goapi.Router'标记字段注释
- path: 路由地址
- method: 访问方法。多个内容用'，'分隔
- summary: 该API的简短摘要。
- desc: API的描述。CommonMark语法可用于富文本表示。
- tags: 多个内容用'，'分隔
### 方法中参数结构标签的标注
- header
  - 可以使用常用类型(ptr, slice)，在切片类型中使用'，'拆分
  - 值是字段的别名，添加'omitempty'则可为空
- cookie
  - 可以使用常用的类型(ptr, slice)或`*http.Cookie'，在切片类型中使用'，'拆分
  - 值是字段的别名，添加'omitempty'则可为空
- query
  - 可以使用常用类型(ptr, slice)
  - 值是字段的别名，添加'omitempty'则可为空
- path
  - 可以使用常用类型(ptr, slice)，在切片类型中使用'，'拆分
  - 值是字段的别名，添加'omitempty'则可为空
- form
  - 可以使用常用类型(ptr, slice)，在切片类型中使用'，'拆分
  - 默认媒体类型'application/x-www-form-urlencoded'，如果有file文件存在媒体类型为'multipart/form-data'
  - 值是字段的别名，添加'omitempty'则可为空
- file
  - 可以使用类型为‘*multipart.FileHeader'或'[]*multipart.FileHeader'
  - 默认媒体类型“multipart/form-data”
  - 值是字段的别名，添加'omitempty'则可为空
- body
  - 固定常用值为'xml'和'json', 多个值用','分割
  - 该值适用于其他媒体类型(例如'text/plain')，值类型为'[]byte'， 'string'或'io.ReadCloser'
  - 值json表示媒体类型'application/json', 值xml表示媒体类型'application/xml'
  - 标签使用值的主体, 添加'omitempty'则可为空
### 结构标签注释
- regexp
    - 值的正则表达式
    - 验证器限制 **字符串** 类型
    - 相当于OpenAPI的 **pattern**
- enum
    - 值的枚举
    - 验证器限制 **整数** **数字** **布尔** **字符串** 类型
    - 逗号分割(,)
- default
    - 默认值
- example
    - 实例值
- desc
    - 字段描述
    - 相当于OpenAPI的 **description**
- lt
    - 小于字段值
    - 验证器限制 **整数** **数字** 类型
    - 相当于OpenAPI的 **exclusiveMaximum**
- lte
    - 小于等于字段值
    - 验证器限制 **整数** **数字** 类型
    - 相当于OpenAPI的 **maximum**
- gt
    - 大于字段值
    - 验证器限制 **整数** **数字** 类型
    - 相当于OpenAPI的 **exclusiveMinimum**
- gte
    - 大于等于字段值
    - 验证器限制 **整数** **数字** 类型
    - 相当于OpenAPI的 **minimum**
- multiple
    - 值的乘数
    - 验证器限制 **整数** **数字** 类型
- max
    - 值的最大长度
    - 验证器限制 **字符串** **数组** **对象** 类型
- min
    - 值的最小长度
    - 验证器限制 **字符串** **数组** **对象** 类型
- unique
    - 验证数组值唯一
    - 验证器限制 **数组** 类型
## 响应注释
### 如果响应是goapi的实现。响应界面，可以设置一些功能
- **HTTPResponse[T]** 可以设置http的code和header
- **FileResponse** 可以返回可下载文件
- **SSEResponse** 可以以服务器发送事件格式返回内容
- **HTMLResponse** 可以返回HTML页面
## 错误对应注释以及用法
~~~go
response.HTTPException(404, "error message")
~~~
- 具体的返回信息可以使用'HTTPExceptionHandler'方法配置
- 第一个参数是HTTP状态码，第二个是错误消息，第三个是返回的报头设置
## 关于
使用类似于Python中的FastAPI的API生成文档