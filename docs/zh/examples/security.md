## [<<](examples.md) 如何使用鉴权验证
- 鉴权在openapi中的 **securitySchemes** 判断以结构体名称和参数字段名称为key
- 相同的名称会覆盖，相同的用法定义的key需一致
- 所有的鉴权接口实现中都可以使用参数 **header**,**cookie**, **query**鉴权
### HTTPBearer鉴权定义
~~~go
// 需要实现接口
type HTTPBearer interface {
	HTTPBearer(token string)
}

type Auth struct {
	Ctx *goapi.Context // 定义该字段可以传递上下文
}

func (a *Auth)HTTPBearer(token string)  {
    // 逻辑处理
	if token != "123456" {
		response.HTTPException(401,"验证失败")
	}
}
~~~
### HTTPBasic鉴权定义
~~~go
// 需要实现接口
type HTTPBasic interface {
	HTTPBasic(username, password string)
}

type Auth struct {
	Ctx *goapi.Context // 定义该字段可以传递上下文
}

func (a *Auth)HTTPBasic(username, password string)  {
	// 逻辑处理 
	if username != "admin" && password != "123456" {
		goapi.HTTPException(401,"验证失败")
	}
}
~~~
### ApiKey鉴权定义
~~~go
// 需要实现接口
type ApiKey interface {
	ApiKey()
}

type Auth struct {
	Ctx *goapi.Context // 定义该字段可以传递上下文
	Token string `header:"Token" desc:"需要验证的token"` // header通用验证
	Name  string `query:"Name" desc:"需要验证的名称"` // query通用验证
	ID int64 `cookie:"ID" desc:"需要验证的ID"` // cookie通用验证
}

func (a *Auth)ApiKey()  {
	// 逻辑处理 
	if a.Token != "admin" && a.Name != "admin" && a.ID != 15 {
		goapi.HTTPException(401,"验证失败")
	}
}
~~~
### HTTPBearerJWT鉴权定义
rsa模式
~~~go
// 需要实现接口
type HTTPBearerJWT interface {
	// EncryptKey the encrypted key
	EncryptKey() (any, error)

	// DecryptKey the Decryption key
	DecryptKey() (any, error)

	// SigningMethod encryption and decryption methods
	SigningMethod() SigningMethod

	// HTTPBearerJWT jwt logical
	HTTPBearerJWT(jwt *JWT)
}

var privateKey, _ = os.ReadFile("private.pem")
var publicKey, _ = os.ReadFile("public.pem")

func (a *Auth) EncryptKey() (any, error) {
	block, _ := pem.Decode(privateKey)
	return x509.ParsePKCS8PrivateKey(block.Bytes)
}

func (a *Auth) DecryptKey() (any, error) {
	block, _ := pem.Decode(publicKey)
	return x509.ParsePKIXPublicKey(block.Bytes)
}

func (a *Auth) SigningMethod() goapi.SigningMethod {
	return goapi.SigningMethodRS256
}

func (a *Auth) HTTPBearerJWT(jwt *goapi.JWT) { 
	// 逻辑处理
	if jwt.ID!="147258" {
		goapi.HTTPException(401,"验证失败")
	}
}
~~~
其他模式
~~~go
// 需要实现接口
type HTTPBearerJWT interface {
	// EncryptKey the encrypted key
	EncryptKey() (any, error)

	// DecryptKey the Decryption key
	DecryptKey() (any, error)

	// SigningMethod encryption and decryption methods
	SigningMethod() SigningMethod

	// HTTPBearerJWT jwt logical
	HTTPBearerJWT(jwt *JWT)
}

var key []byte("1234568547854525")

func (a *Auth) EncryptKey() (any, error) {
	return key,nil
}

func (a *Auth) DecryptKey() (any, error) {
	return key,nil
}

func (a *Auth) SigningMethod() goapi.SigningMethod {
	return goapi.SigningMethodHS256
}

func (a *Auth) HTTPBearerJWT(jwt *goapi.JWT) { 
	// 逻辑处理
	if jwt.ID!="147258" {
		goapi.HTTPException(401,"验证失败")
	}
}
~~~
鉴权使用
~~~go
func (*index)Index(input struct{
	router goapi.Router `path:"/index" method:"GET"`
	Auth *Auth   // 添加该行则该路由使用这个鉴权
})  {
	
}
~~~