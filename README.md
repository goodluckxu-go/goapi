# goapi
Using the HTTP framework of OpenAPI3.1 documentation

English | [中文](README_zh.md)

## usage
~~~bash
go get github.com/goodluckxu-go/goapi
~~~
main.go
~~~go
import (
	"github.com/fatih/color"
	"github.com/goodluckxu-go/goapi"
)

func main() {
	color.NoColor = true // Turn off the default logging console color
	api := goapi.GoAPI(true, "/docs")
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


// Implement ApiKey interface
type AdminAuth struct {
	Token  string   `header:"Token"`
	Admin  string          // Define a value and retrieve it from the controller
}

func (h *AdminAuth) ApiKey() {
	if h.Token != "123456" {
		response.HTTPException(401, "token is error")
	}
	h.Admin = "admin"
}
~~~
## Verify multilingual Settings
You can implement the 'goapi.Lang' interface yourself
~~~go
api.SetLang(&lang.ZhCn{}) // Default 'EnUs' English comments
~~~
## Log output setting
Set before initializing the api
~~~go
goapi.SetLogLevel(goapi.LogInfo | goapi.LogWarning)
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
  - The value is for other media types(example 'text/plain'), The type is '[]byte', 'string' or 'io.ReadCloser'
  - Value of json is media type 'application/json', xml is media type 'application/xml'
  - Body of tag use value, 'omitempty' is nullable
  - When the value is 'application/octet-stream', indicates that the body is uploaded as a file
### Structure tag annotation
- regexp
    - Regular expression of value
    - Validator limit **string** type
    - Equivalent to OpenAPI **pattern**
- enum
    - Enumeration of values
    - Validator limit **integer** **number** **boolean** **string** type
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
    - Validator limit **integer** **number** type
    - Equivalent to OpenAPI **exclusiveMaximum**
- lte
    - Less than or equal to value
    - Validator limit **integer** **number** type
    - Equivalent to OpenAPI **maximum**
- gt
    - Greater than value
    - Validator limit **integer** **number** type
    - Equivalent to OpenAPI **exclusiveMinimum**
- gte
    - Greater than or equal to value
    - Validator limit **integer** **number** type
    - Equivalent to OpenAPI **minimum**
- multiple
    - Multipliers of values
    - Validator limit **integer** **number** type
- max
    - The maximum length of the value
    - Validator limit **string** **array** **object** type
- min
    - The minimum length of the value
    - Validator limit **string** **array** **object** type
- unique
    - The value of the array is unique
    - Validator limit **array** type
## Response annotation
### if response is an implementation of the goapi.Response interface, you can set some functions
- **HTTPResponse[T]** can set httpCode, header, cookie. Content-Type value is 'application/json','application/xml'
- **FileResponse** can return downloadable files. Content-Type value is 'application/octet-stream'
- **SSEResponse** can return content in Server Sent Events format. Content-Type value is 'text/event-stream'
- **HTMLResponse** can return to HTML page. Content-Type value is 'text/html'
- **TextResponse** return in text mode, can set header, cookie. Content-Type default value is 'text/plain',resettable Content-Type
## Error corresponding comment
~~~go
response.HTTPException(404, "error message")
~~~
- Specific return information can be configured using the 'HTTPExceptionHandler' method
- The first parameter is the HTTP status code, the second is the error message, and the third is the header setting returned
## About
Generate documentation using an API similar to FastAPI in Python