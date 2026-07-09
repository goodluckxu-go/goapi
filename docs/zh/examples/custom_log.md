## [<<](examples.md) 如何自定义日志
### 自定义示例
~~~go
// 实现接口
type Logger interface {
	Debug(format string, a ...any)
	Info(format string, a ...any)
	Warning(format string, a ...any)
	Error(format string, a ...any)
	Fatal(format string, a ...any)
}

type CustomLog struct {
	ctx *Context
}

func (*CustomLog) Debug(format string, a ...any) {
	
}

func (c *CustomLog) Info(format string, a ...any) {
	if c.ctx != nil {
		_ = c.ctx.RequestID // 可获取每次请求的ID
		_ = c.ctx.ChildPath // 可获取每次请求的程序定义请求前缀
	}
}

func (*CustomLog) Warning(format string, a ...any) {

}

func (*CustomLog) Error(format string, a ...any) {

}

func (*CustomLog) Fatal(format string, a ...any) {

}
~~~
### 日志写入传入上下文判断
~~~go
// 实现接口
type LoggerWithContext interface {
	WithContext(ctx *Context) Logger
}

// 实现接口
func (c *CustomLog) WithContext(ctx *Context) Logger {
	return &CustomLog{ctx: ctx}
}
~~~
### 使用日志
~~~go
func main() {
	api := goapi.Default(true)
	api.GenerateRequestID = true // 生成每次请求的ID
	api.UseXRequestIDHeader = true // 优先使用请求头X-Request-ID，并写回响应头
	api.SetLogger(&CustomLog{})
}
~~~
开启`UseXRequestIDHeader`后，如果请求头中存在`X-Request-ID`，会优先作为`ctx.RequestID`使用；如果不存在，则自动生成新的请求ID，并通过响应头`X-Request-ID`返回。
