## [<<](examples.md) Context方法详解
### Deadline() (deadline time.Time, ok bool)
继承http.Request.Context().Deadline()方法
### Done() <-chan struct{}
继承http.Request.Context().Done()方法
### Err() error
继承http.Request.Context().Err()方法
### Set(key string, value any)
设置一个参数用于上下文传递
### Get(key string) (value any, ok bool)
获取一个上下文设置的参数
### Value(key any) any
获取一个上下文设置的参数，兼容context.Context类的Value方法
### FullPath() string
获取全路由方法，例如：/user/{id}
### Next()
中间件执行逻辑，在中间件中必须使用，否则无法执行下一步
### Logger() Logger
获取日志，该日志继承上下文处理，可设置GenerateRequestID = true后合并每次请求的所有日志
### RemoteIP() string
获取的客户端的IP，没有经过转发的，一般用于局域网获取真实客户端IP地址
### ClientIP() string
获取的客户端的IP，可获取转发的X-Forwarded-For和X-Real-IP的header头信息
### Query() url.Values
获取所有的query参数集合，该集合已经做缓存
### Redirect(status int, location string)
- 跳转重定向状态和地址
- 状态码定义

| HTTP 状态码 |   类型   |  等幂性保持  |  缓存性  |         典型场景                   | 官方/常见定义 |
|------------|---------|------------|---------|-----------------------------------|------------|
| 301        |永久重定向 | 可能改为GET  | 可缓存   | 域名迁移，HTTP升级为HTTPS            | RFC 7231   |
| 302        |临时重定向 | 可能改为GET  | 不缓存   | 临时跳转（传统用法）                  | RFC 7231   |
| 303        |特殊/其他 | 强制改为GET  | 禁止缓存  | 表单提交后跳转，防止重复提交           | RFC 7231   |
| 307        |临时重定向 | 方法保持不变  | 不缓存   | 临时跳转，需保留请求方法（如表单提交）   | RFC 7231   |
| 308        |永久重定向 | 方法保持不变  | 可缓存   | 永久迁移，需保留请求方法的场景          | RFC 7231   |