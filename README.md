# goapi
使用OpenAPI3.1文档的HTTP框架

使用说明:
- [中文文档](docs/zh/index.md)
## 用法
~~~bash
go get github.com/goodluckxu-go/goapi/v2
~~~
## 功能
- 实现了http服务，路由使用gin路由模式的前缀树方式实现
- 集成swagger+openapi3.1.0文档的访问
- 实现了openapi中验证和goapi程序验证的同步
- 实现了自定义中间件
- 实现了鉴权认证
- 实现了路由组模式
- 实现了多个程序模块组的模式
- ......
## 关于
使用类似于Python中的FastAPI的API生成文档