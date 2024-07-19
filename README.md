# goapi
Using the HTTP framework of OpenAPI3.1 documentation

## usage
~~~bash
go get github.com/goodluckxu-go/goapi
~~~
main.go
~~~go
import (
	"github.com/fatih/color"
	"github.com/goodluckxu-go/goapi"
	"github.com/goodluckxu-go/goapi/app"
)

func main() {
	color.NoColor = true // Turn off console color, default color
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
user_controller.go
~~~go
type UserController struct {
}

type UserListRouter struct {
}

type Req struct {
	Page int `json:"page" desc:"now page" gt:"0"`
	Limit int `json:"page" desc:"limit a page count" gte:"10" lte:"100"`
}

func (u *UserListRouter) Index(input struct {
	router  goapi.Router `path:"/index/{id}" method:"put" summary:"test api" desc:"test api" tags:"admin"`
	Auth    *AdminAuth
	ID      string `path:"id" regexp:"^\d+$"` // path 
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
		goapi.HTTPException(401, "token is error")   
	}
	h.Admin = "admin"
}

// Implement HTTPBasic interface
type AdminAuth struct {
	Admin  string          // Define a value and retrieve it from the controller
}

func (h *AdminAuth) HTTPBasic(username,password string) {
	if username != "admin" || password != "123456" {
		goapi.HTTPException(401, "token is error")
	} 
	h.Admin = "admin"
}

// Auth verification
type UserListHeader struct { 
	Token     string
}

// Implement ApiKey interface
type AdminAuth struct {
	Token  string   `header:"Token"`
	Admin  string          // Define a value and retrieve it from the controller
}

func (h *AdminAuth) ApiKey() {
	if h.Token != "123456" {
		goapi.HTTPException(401, "token is error")
	}
	h.Admin = "admin"
}
~~~
### 'goapi.Router' tag field annotation
- path: Access Routing
- method: Access method. Multiple contents separated by ','
- summary: A short summary of the API.
- desc: A description of the API. CommonMark syntax MAY be used for rich text representation.
- tags: Multiple contents separated by ','
### Annotation of parameter structure tag in the method
- header
  - Can use commonly used types(ptr, slice), in slice type use ',' split
  - Value is an alias for a field, 'omitempty' is nullable
- cookie
  - Can use commonly used types(ptr, slice) or '*http.Cookie', in slice type use ',' split
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
  - Can use commonly used '*multipart.FileHeader' or '[]*multipart.FileHeader'
  - default media type 'multipart/form-data'
  - Value is an alias for a field, 'omitempty' is nullable
- body
  - The values are 'xml' and 'json', Multiple uses ',' segmentation
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
## Response annotation
- if response is an implementation of the goapi.Response interface, you can set the http Code and header
- else do not set the header, and set the HTTP code to 200
## Error corresponding comment
~~~go
goapi.HTTPException(404, "error message")
~~~
- Specific return information can be configured using the 'HTTPExceptionHandler' method
- The first parameter is the HTTP status code, the second is the error message, and the third is the header setting returned
## About
Generate documentation using an API similar to FastAPI in Python