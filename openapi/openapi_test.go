package openapi

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

var jsonStr = `{"components":{"securitySchemes":{"httpBasic":{"scheme":"basic","type":"http"}}},"info":{"title":"GoAPI","version":"1.0.0"},"openapi":"3.1.0","paths":{"/user/{id}":{"description":"user handle","get":{"description":"user info","operationId":"/user/{id}_get","parameters":[{"description":"pk","in":"path","name":"id","required":true,"schema":{"type":["integer"]}},{"description":"type","in":"query","name":"type","schema":{"type":["string"]}}],"responses":{"default":{"content":{"application/json":{"schema":{"description":"content","properties":{"age":{"type":["integer"]},"id":{"type":["integer"]},"name":{"type":["string"]}},"title":"content","type":["object"]}}},"description":"desc","headers":{"Set-Token":{"description":"set token","required":false,"schema":{"type":["string"]}}},"links":{"bd":{"description":"baidu link","operationRef":"https://www.baidu.com","parameters":{"id":"1"},"requestBody":"test"}}}},"summary":"user info","tags":["admin"]},"put":{"callbacks":{"callback":{"{$request.query.callbackUrl}":{"description":"callback","post":{"description":"callback","operationId":"callback_post","parameters":[{"description":"type","in":"query","name":"callbackUrl","required":true,"schema":{"type":["string"]}}],"requestBody":{"$ref":"#/paths/~1user~1%7Bid%7D/put/requestBody"},"responses":{"default":{"$ref":"#/paths/~1user~1%7Bid%7D/put/responses/default"}},"summary":"callback","tags":["admin"]},"summary":"callback"}}},"description":"edit user","operationId":"/user/{id}_put","parameters":[{"description":"pk","in":"path","name":"id","required":true,"schema":{"type":["integer"]}}],"requestBody":{"content":{"application/json":{"schema":{"$ref":"#/paths/~1user~1%7Bid%7D/get/responses/default/content/application~1json/schema"}}},"description":"set body"},"responses":{"default":{"content":{"application/json":{"schema":{"type":["boolean"]}}},"description":"aaa"}},"summary":"edit user","tags":["admin"]},"summary":"user handle"}},"security":[{"httpBasic":[]}],"tags":[{"description":"admin manager","name":"admin"}]}`

func TestMarshalJSON(t *testing.T) {
	callback := &Callback{}
	callback.Set("{$request.query.callbackUrl}", &PathItem{
		Summary:     "callback",
		Description: "callback",
		Post: &Operation{
			Tags:        []string{"admin"},
			Summary:     "callback",
			Description: "callback",
			OperationId: "callback_post",
			Parameters: []*Parameter{
				{
					Name:        "callbackUrl",
					In:          "query",
					Description: "type",
					Required:    true,
					Schema: &Schema{
						Type: []string{"string"},
					},
				},
			},
			RequestBody: &RequestBody{
				Ref: "#/paths/~1user~1%7Bid%7D/put/requestBody",
			},
			Responses: &Responses{
				Default: &Response{
					Ref: "#/paths/~1user~1%7Bid%7D/put/responses/default",
				},
			},
		},
	})
	paths := &Paths{}
	paths.Set("/user/{id}", &PathItem{
		Summary:     "user handle",
		Description: "user handle",
		Get: &Operation{
			Tags:        []string{"admin"},
			Summary:     "user info",
			Description: "user info",
			OperationId: "/user/{id}_get",
			Parameters: []*Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "pk",
					Required:    true,
					Schema: &Schema{
						Type: []string{"integer"},
					},
				},
				{
					Name:        "type",
					In:          "query",
					Description: "type",
					Schema: &Schema{
						Type: []string{"string"},
					},
				},
			},
			Responses: &Responses{
				Default: &Response{
					Description: "desc",
					Headers: map[string]*Header{
						"Set-Token": {
							Description: "set token",
							Schema: &Schema{
								Type: []string{"string"},
							},
						},
					},
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{
								Type:        []string{"object"},
								Title:       "content",
								Description: "content",
								Properties: map[string]*Schema{
									"id":   {Type: []string{"integer"}},
									"name": {Type: []string{"string"}},
									"age":  {Type: []string{"integer"}},
								},
							},
						},
					},
					Links: map[string]*Link{
						"bd": {
							OperationRef: "https://www.baidu.com",
							Parameters: map[string]any{
								"id": "1",
							},
							RequestBody: "test",
							Description: "baidu link",
						},
					},
				},
			},
		},
		Put: &Operation{
			Tags:        []string{"admin"},
			Summary:     "edit user",
			Description: "edit user",
			OperationId: "/user/{id}_put",
			Parameters: []*Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "pk",
					Required:    true,
					Schema: &Schema{
						Type: []string{"integer"},
					},
				},
			},
			RequestBody: &RequestBody{
				Description: "set body",
				Content: map[string]*MediaType{
					"application/json": {
						Schema: &Schema{
							Ref: "#/paths/~1user~1%7Bid%7D/get/responses/default/content/application~1json/schema",
						},
					},
				},
			},
			Responses: &Responses{
				Default: &Response{
					Description: "aaa",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{
								Type: []string{"boolean"},
							},
						},
					},
				},
			},
			Callbacks: map[string]*Callback{
				"callback": callback,
			},
		},
	})
	openapi := &OpenAPI{
		OpenAPI: "3.1.0",
		Info: &Info{
			Title:   "GoAPI",
			Version: "1.0.0",
		},
		Paths: paths,
		Components: &Components{
			SecuritySchemes: map[string]*SecurityScheme{
				"httpBasic": {
					Type:   "http",
					Scheme: "basic",
				},
			},
		},
		Security: []*SecurityRequirement{
			{"httpBasic": []string{}},
		},
		Tags: []*Tag{
			{Name: "admin", Description: "admin manager"},
		},
	}
	buf, _ := json.Marshal(openapi)
	assert.JSONEq(t, jsonStr, string(buf))
}

func TestUnmarshalJSON(t *testing.T) {
	callback := &Callback{}
	callback.Set("{$request.query.callbackUrl}", &PathItem{
		Summary:     "callback",
		Description: "callback",
		Post: &Operation{
			Tags:        []string{"admin"},
			Summary:     "callback",
			Description: "callback",
			OperationId: "callback_post",
			Parameters: []*Parameter{
				{
					Name:        "callbackUrl",
					In:          "query",
					Description: "type",
					Required:    true,
					Schema: &Schema{
						Type: []string{"string"},
					},
				},
			},
			RequestBody: &RequestBody{
				Ref: "#/paths/~1user~1%7Bid%7D/put/requestBody",
			},
			Responses: &Responses{
				Default: &Response{
					Ref: "#/paths/~1user~1%7Bid%7D/put/responses/default",
				},
			},
		},
	})
	paths := &Paths{}
	paths.Set("/user/{id}", &PathItem{
		Summary:     "user handle",
		Description: "user handle",
		Get: &Operation{
			Tags:        []string{"admin"},
			Summary:     "user info",
			Description: "user info",
			OperationId: "/user/{id}_get",
			Parameters: []*Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "pk",
					Required:    true,
					Schema: &Schema{
						Type: []string{"integer"},
					},
				},
				{
					Name:        "type",
					In:          "query",
					Description: "type",
					Schema: &Schema{
						Type: []string{"string"},
					},
				},
			},
			Responses: &Responses{
				Default: &Response{
					Description: "desc",
					Headers: map[string]*Header{
						"Set-Token": {
							Description: "set token",
							Schema: &Schema{
								Type: []string{"string"},
							},
						},
					},
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{
								Type:        []string{"object"},
								Title:       "content",
								Description: "content",
								Properties: map[string]*Schema{
									"id":   {Type: []string{"integer"}},
									"name": {Type: []string{"string"}},
									"age":  {Type: []string{"integer"}},
								},
							},
						},
					},
					Links: map[string]*Link{
						"bd": {
							OperationRef: "https://www.baidu.com",
							Parameters: map[string]any{
								"id": "1",
							},
							RequestBody: "test",
							Description: "baidu link",
						},
					},
				},
			},
		},
		Put: &Operation{
			Tags:        []string{"admin"},
			Summary:     "edit user",
			Description: "edit user",
			OperationId: "/user/{id}_put",
			Parameters: []*Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "pk",
					Required:    true,
					Schema: &Schema{
						Type: []string{"integer"},
					},
				},
			},
			RequestBody: &RequestBody{
				Description: "set body",
				Content: map[string]*MediaType{
					"application/json": {
						Schema: &Schema{
							Ref: "#/paths/~1user~1%7Bid%7D/get/responses/default/content/application~1json/schema",
						},
					},
				},
			},
			Responses: &Responses{
				Default: &Response{
					Description: "aaa",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: &Schema{
								Type: []string{"boolean"},
							},
						},
					},
				},
			},
			Callbacks: map[string]*Callback{
				"callback": callback,
			},
		},
	})
	openapi := &OpenAPI{
		OpenAPI: "3.1.0",
		Info: &Info{
			Title:   "GoAPI",
			Version: "1.0.0",
		},
		Paths: paths,
		Components: &Components{
			SecuritySchemes: map[string]*SecurityScheme{
				"httpBasic": {
					Type:   "http",
					Scheme: "basic",
				},
			},
		},
		Security: []*SecurityRequirement{
			{"httpBasic": []string{}},
		},
		Tags: []*Tag{
			{Name: "admin", Description: "admin manager"},
		},
	}
	api := &OpenAPI{}
	_ = json.Unmarshal([]byte(jsonStr), api)
	inBuf, _ := json.Marshal(openapi)
	outBuf, _ := json.Marshal(api)
	assert.JSONEq(t, string(inBuf), string(outBuf))
}
