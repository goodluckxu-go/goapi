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

type CustomLog struct{}

func (*CustomLog) Debug(format string, a ...any) {
	
}

func (*CustomLog) Info(format string, a ...any) {

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
type LoggerContext interface {
	SetContext(ctx *Context)
}

// 实现接口
func (c *CustomLog) SetContext(ctx *Context) {
    _ = ctx.RequestID // 可获取每次请求的ID
    _ = ctx.ChildPath // 可获取每次请求的程序定义请求前缀
}
~~~
### 使用日志
~~~go
func main() {
	api := goapi.Default(true)
	api.GenerateRequestID = true // 生成每次请求的ID
	api.SetLogger(&CustomLog{})
}
~~~