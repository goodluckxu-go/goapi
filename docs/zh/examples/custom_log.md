## [<<](examples.md) 如何自定义日志
### 自定义示例
~~~go
// 实现接口
type Logger interface {
	// Debug 输出调试日志，用于开发或排查问题。
	Debug(format string, a ...any)
	// Info 输出普通运行日志。
	Info(format string, a ...any)
	// Warning 输出警告日志，不会中断请求处理。
	Warning(format string, a ...any)
	// Error 输出错误日志，用于记录失败操作。
	Error(format string, a ...any)
	// Fatal 输出严重级别日志，但不能调用 os.Exit、panic 或让程序退出。
	Fatal(format string, a ...any)
	// WithFields 返回带结构化字段的日志实例。
	WithFields(keysAndValues ...any) Logger
}

type CustomLog struct {
	ctx    *Context
	Fields []LogField
}

func (c *CustomLog) defaultLogger() *DefaultLogger {
	return &DefaultLogger{Fields: c.Fields}
}

func (c *CustomLog) Debug(format string, a ...any) {
	c.defaultLogger().Debug(format, a...)
}

func (c *CustomLog) Info(format string, a ...any) {
	if c.ctx != nil {
		_ = c.ctx.RequestID // 可获取每次请求的ID
		_ = c.ctx.ChildPath // 可获取每次请求的程序定义请求前缀
	}
	c.defaultLogger().Info(format, a...)
}

func (c *CustomLog) Warning(format string, a ...any) {
	c.defaultLogger().Warning(format, a...)
}

func (c *CustomLog) Error(format string, a ...any) {
	c.defaultLogger().Error(format, a...)
}

func (c *CustomLog) Fatal(format string, a ...any) {
	c.defaultLogger().Fatal(format, a...)
}

func (c *CustomLog) WithFields(keysAndValues ...any) Logger {
	if len(keysAndValues) == 0 {
		return c
	}
	fields := ParseLogFields(keysAndValues...)
	newFields := make([]LogField, len(c.Fields)+len(fields))
	n := copy(newFields, c.Fields)
	copy(newFields[n:], fields)
	return &CustomLog{ctx: c.ctx, Fields: newFields}
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
	log := &CustomLog{ctx: ctx, Fields: c.Fields}
	if ctx.RequestID != "" {
		return log.WithFields("request_id", ctx.RequestID)
	}
	return log
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

### 日志展示
~~~text
INFO      [2026-07-10 10:00:00] GoAPI running on http://:8080 (Press CTRL+C to quit)
INFO      [2026-07-10 10:00:01] [1.234ms] 127.0.0.1 - "GET /index" 200 OK [request_id=req-001]
FATAL     [2026-07-10 10:00:03] panic: unexpected value [recovered] [request_id=req-001]
~~~
`Fatal`只表示严重级别日志，日志实现不能在该方法中退出程序；框架会继续执行异常处理逻辑并返回错误响应。
