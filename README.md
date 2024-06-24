# goapi
About the OpenAPI3.1.0 usage of http

## usage
~~~bash
go get github.com/goodluckxu-go/goapi
~~~
main.go入口文件
~~~go
import (
"github.com/goodluckxu-go/goapi"
"github.com/goodluckxu-go/goapi/app"
)

func main() {
	api := goapi.GoAPI(&app.Gin{}, true, "/docs")
	api.SetResponseMediaType(goapi.JSON)
	api.HTTPExceptionHandler(func(httpCode int, detail string) goapi.Response {
		return &goapi.HTTPResponse[Error]{
			HttpCode: httpCode, 
			Body: Error{
				Code:  httpCode, 
				Error: detail,
			},
		}
	})
	api.IncludeRouter(&IndexController{}, "/v1", true, func(ctx *goapi.Context) {
		ctx.Next()
	})
	_ = api.Run("127.0.0.1:8080")
}
~~~
user_controller.go控制器文件
~~~go
type UserController struct {
}

type UserListRouter struct {
}

func (u *UserListRouter) Index(input struct {
	router  goapi.Router `path:"/index/{id}" method:"put" summary:"test api" desc:"test api" tags:"admin"`
	Auth    *AdminAuth
	ID      string `path:"id"` // path 
	Req     Req `body:"json"`
}) Resp {
	return Resp{}
}

// Implement HTTPBearer interface
type AdminAuth struct {
	Admin  string          // Define a value and retrieve it from the controller
}

func (h *AdminAuth) HTTPBearer(token string) {
	if token != "123456" {
		response.HTTPException(401, "token is error")   
	}
	h.Admin = "admin"
}

// Implement HTTPBasic interface
type AdminAuth struct {
	Admin  string          // Define a value and retrieve it from the controller
}

func (h *AdminAuth) HTTPBasic(username,password string) {
	if username != "admin" || password != "123456" {
		response.HTTPException(401, "token is error")
	} 
	h.Admin = "admin"
}

// Auth verification
type UserListHeader struct { 
	Token     string
	param.Header     // inherit
}

// Implement ApiKey interface
type AdminAuth struct {
	Header *UserListHeader
	Admin  string          // Define a value and retrieve it from the controller
}

func (h *AdminAuth) ApiKey() {
	if h.Header.token != "123456" {
		response.HTTPException(401, "token is error")
	}
	h.Admin = "admin"
}
~~~

### Structure tag field annotation
- header
  - Can use commonly used types(ptr, slice), in slice type use ',' split
  - Value is an alias for a field, 'omitempty' is nullable
- cookie
  - Can use commonly used types(ptr, slice) or *http.Cookie, in slice type use ',' split
  - Value is an alias for a field, 'omitempty' is nullable
- query
  - Can use commonly used types(ptr, slice)
  - Value is an alias for a field, 'omitempty' is nullable
- path
  - Can use commonly used types(ptr, slice), in slice type use ',' split
  - Value is an alias for a field, 'omitempty' is nullable
- form
  - Can use commonly used types(ptr, slice), in slice type use ',' split
  - default media type 'application/x-www-form-urlencoded', if file exists 'multipart/form-data'
  - Value is an alias for a field, 'omitempty' is nullable
- file
  - Can use commonly used *multipart.FileHeader or []*multipart.FileHeader
  - default media type 'multipart/form-data'
  - Value is an alias for a field, 'omitempty' is nullable
- body
  - The values are xml and json, Multiple uses ',' segmentation
  - Value of json is media type 'application/json', xml is media type 'application/xml'
  - Body of tag use value, 'omitempty' is nullable
### Structure tag annotation
- regexp
    - Regular expression of value
    - Equivalent to OpenAPI **pattern**
- enum
    - Enumeration of values
    - Limit **integer** **number** **boolean** **string** type
    - Comma division (**,**)
- default
    - Default value
- example
    - Example value
- desc
    - Field description
    - Equivalent to OpenAPI **description**
- lt
    - Less than value
    - Limit **integer** **number** type
    - Equivalent to OpenAPI **exclusiveMaximum**
- lte
    - Less than or equal to value
    - Limit **integer** **number** type
    - Equivalent to OpenAPI **maximum**
- gt
    - Greater than value
    - Limit **integer** **number** type
    - Equivalent to OpenAPI **exclusiveMinimum**
- gte
    - Greater than or equal to value
    - Limit **integer** **number** type
    - Equivalent to OpenAPI **minimum**
- multiple
    - Multipliers of values
    - Limit **integer** **number** type
- max
    - The maximum length of the value
    - Limit **string** **array** **object** type
- min
    - The minimum length of the value
    - Limit **string** **array** **object** type
- unique
    - The value of the array is unique
    - Limit **array** type
## About
Generate documentation using an API similar to FastAPI in Python