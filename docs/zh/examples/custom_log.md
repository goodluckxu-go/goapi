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
### 日志添加每次记录请求id
~~~go
// 实现接口
type LoggerRequestID interface {
	SetRequestID(id string)
}

// 上面的所有请求都可以用c.id判断是哪次请求
func (c *CustomLog) SetRequestID(id string) {
    c.id = id
}
~~~
### 使用日志
~~~go
func main() {
	// 设置日志级别
	goapi.SetLogLevel(goapi.LogInfo | goapi.LogError | goapi.LogDebug | goapi.LogFail | goapi.LogError)
	api := goapi.GoAPI(true)
	api.SetLogger(&CustomLog{})
}
~~~