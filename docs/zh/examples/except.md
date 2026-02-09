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
	api := goapi.GoAPI(true)
	api.HTTPException(func(httpCode int, detail string) any {
		return MainError{
			Code: httpCode,
			Error:  detail,
		}
	})
	admin:=api.Child("/admin", "/admin")
	{
		admin.HTTPException(func(httpCode int, detail string) any {
			return ChildError{
				Code: httpCode, 
				Msg:  detail,
			}
		})
		admin.AddMiddleware(nil)
	}
	user:=api.Child("/user", "/v1")
	{
		user.HTTPException(func(httpCode int, detail string) any {
			return ChildError{
				Code: httpCode, 
				Msg:  detail,
			}
		})
		user.AddMiddleware(nil)
	}
}
~~~
### 使用通用错误
~~~go
goapi.HTTPException(404, "错误信息")
~~~