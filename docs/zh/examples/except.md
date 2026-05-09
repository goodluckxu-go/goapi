## [<<](examples.md) 如何处理异常或错误返回
多个子模块单独定义
### 设置通用错误返回格式
~~~go
type MainError struct {
	Code int
	Error string
}

func(m MainError)GetStatus() int {
	return m.Code
}

func(m MainError)GetHeader() http.Header {
	return http.Header{
		"Content-Type": {"text/html; charset=utf-8"}
	}
}

func(m MainError)GetBody()any {
	return m.Error
}

type ChildError struct {
	Code int
	Msg  string
}

func main() {
	api := goapi.Default(true)
	api.HTTPError(func(err error) any {
		switch val := err.(type) {
		case *goapi.HTTPError:
			// 使用goapi.NewHTTPError方法返回的
			return MainError{Code: val.Code, Error: val.Message}
		case ...:
			// 支持自定义返回格式，通用处理在这里处理
		}
		// 其他通用错误返回
		return MainError{Code: -1, Error: err.Error()}
	})
	admin:=api.Child("/admin", "/admin")
	{
		admin.HTTPError(func(err error) any {
			switch val := err.(type) {
			case *goapi.HTTPError: 
				// 使用goapi.NewHTTPError方法返回的 
				return MainError{Code: val.Code, Error: val.Message}
			case ...: 
				// 支持自定义返回格式，通用处理在这里处理
			}
			// 其他通用错误返回 
			return MainError{Code: -1, Error: err.Error()}
		})
		admin.AddMiddleware(nil)
	}
	user:=api.Child("/user", "/v1")
	{
		user.HTTPError(func(err error) any {
			switch val := err.(type) {
			case *goapi.HTTPError: 
				// 使用goapi.NewHTTPError方法返回的 
				return MainError{Code: val.Code, Error: val.Message}
			case ...: 
				// 支持自定义返回格式，通用处理在这里处理
			}
			// 其他通用错误返回 
			return MainError{Code: -1, Error: err.Error()}
		})
		user.AddMiddleware(nil)
	}
}
~~~
### 使用通用错误
在方法中使用
~~~go
func (*Index) Param(input struct {
	input  goapi.Router `paths:"/param" methods:"GET" summary:"html返回"`
}) (*response.Html,error) {
	params := map[string]any{
		"name": "张三",
	}
	//return nil, goapi.NewHTTPError(401,"测试错误") // 错误处理
	//return response.ReturnHtmlByFile("1.html", params), nil
	return response.ReturnHtml("<div>{{.name}}</div>", nil),nil
}
~~~
security中使用
~~~go
type Auth struct {
	Ctx *goapi.Context // 定义该字段可以传递上下文
}

func (a *Auth)HTTPBearer(token string) error  {
	// 逻辑处理
	if token != "123456" {
		return goapi.NewHTTPError(401,"验证失败")
	}
	return nil
}
~~~