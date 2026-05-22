## [<<](examples.md) 如何添加扩展参数
- 添加 **x-** 可用于扩展参数用于swagger展示
- 可以用上下文 ***goapi.Context** 读取参数的扩展
### 添加扩展用于swagger展示
- 参数类型为 **path** , **query** , **header** , **cookie** , **form** , **file** , **body**
- 可以展示在输入和返回的所有结构体中
- **x-test1**, **x-test2** 会注册到openapi.json里面的Extensions扩展中
- 不能定义 **x-match** 扩展，此为系统扩展，用于标注 **{path:*}** 路由参数匹配
~~~go
func main() {
	api := goapi.Default(true)
	api.Swagger.ShowExtensions = true // 必须配置才能展示，默认false
}

func (*Index) Index(ctx *goapi.Context, input struct{
	router goapi.Router `paths:"/post" methods:"POST"`
	Body BodyReq `body:"json" x-test1:"测试body"`
}) {
}

type BodyReq struct {
	Name        string
	Friends     []UserInfo
	FriendMaps  map[string]UserInfo
}

type UserInfo struct {
	Name string `x-test2:"测试name"`
}
~~~
### 用于上下文 ***goapi.Context** 读取参数的扩展
-只能在go.Router这一层级的可以获取，如下：
~~~go
func (*Index) Index(ctx *goapi.Context, input struct{
	router goapi.Router `paths:"/" methods:"get" x-test1:"1"` // x-test1 支持
	Body struct {
		Page int `query:"page" x-error:"error"` // x-error 不支持, 不和goapi.Router一个层级
	} `x-test2:"2"` // x-test2 支持
	Limit int `query:"limit" x-test3:"3"` // x-test3 支持
}) {
	ctx.Extensions.Get("x-test1")
	ctx.Extensions.Get("x-test2")
	ctx.Extensions.Get("x-test3")
}

type Auth struct {
	Ctx *goapi.Context
}

func (a *Auth) HTTPBearer(token string) {
	a.Ctx.Extensions.Get("x-test1")
	a.Ctx.Extensions.Get("x-test2")
	a.Ctx.Extensions.Get("x-test3")
}
~~~