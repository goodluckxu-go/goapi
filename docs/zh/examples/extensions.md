## [<<](examples.md) 如何添加扩展参数
- 添加 **x-** 扩展参数用于swagger展示
- 可以用上下文 ***goapi.Context** 读取参数的扩展
- 参数类型为 **path** , **query** , **header** , **cookie** , **form** , **file** , **body**
- 不能定义 **x-match** 扩展，此为系统扩展，用于标注 **{path:*}** 路由参数匹配
~~~go
func (*Index) Index(ctx *goapi.Context, input struct{
	router goapi.Router `paths:"/post" methods:"POST"`
	Body BodyReq `body:"json" x-test1:"测试body"`
	Auth *Auth `x-auth:"测试代码"`
}) {
	// 获取x-test1扩展值
	valAny,ok:=ctx.Extensions.Struct("Body").Get("x-test1")
	// 获取x-test2扩展值
	valString:=ctx.Extensions.Struct("Body").Struct("Friends").Slice().Struct("Name").GetString("x-test2")
	valString:=ctx.Extensions.Struct("Body").Struct("FriendMaps").Map().Struct("Name").GetString("x-test2")
    // 获取x-auth
	valString:=ctx.Extensions.Struct("Auth").GetString("x-auth")
}

type BodyReq struct {
	Name        string
	Friends     []UserInfo
	FriendMaps  map[string]UserInfo
}

type UserInfo struct {
	Name string `x-test2:"测试name"`
}

type Auth struct {
	Ctx *goapi.Context
}

func (a *Auth) HTTPBearer(token string) {
	// 运行方法前会添加 ctx.Extensions = ctx.Extensions.Struct("Auth") 
	// 省略了当前鉴权的结构体名称Auth，获取x-auth如下 
	valString:=ctx.Extensions.GetString("x-auth")
	// 如果需要完整的获取则需要添加Root方法,Root方法也可获取鉴权结构体之外的值
	valString:=ctx.Extensions.Root().Struct("Auth").GetString("x-auth")
}
~~~