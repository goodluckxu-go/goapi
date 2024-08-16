package goapi

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkOneRouter(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/index", nil)
	if err != nil {
		panic(err)
	}
	writer := httptest.NewRecorder()
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkOneReturnRouter(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/index/return", nil)
	if err != nil {
		panic(err)
	}
	writer := httptest.NewRecorder()
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkMiddlewareRouter(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/middleware", nil)
	if err != nil {
		panic(err)
	}
	writer := httptest.NewRecorder()
	hd := testGetApiHandler(func(ctx *Context) {
		ctx.Request.Header.Set("Token", "111")
		ctx.Next()
	}, func(ctx *Context) {
		ctx.Request.Header.Set("Token2", "258")
		ctx.Next()
	})
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkPostDataRouter(b *testing.B) {
	buf, _ := json.Marshal(map[string]interface{}{
		"Id":   15,
		"Name": "zs",
	})
	writer := httptest.NewRecorder()
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(http.MethodPost, "/post?type=125", io.NopCloser(bytes.NewBuffer(buf)))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Token", "123")
		req.Header.Set("Cookie", "Projectid=147")
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkPostFileRouter(b *testing.B) {
	ctype, buf := createFile("./README.md")
	writer := httptest.NewRecorder()
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(http.MethodPost, "/post/file", io.NopCloser(bytes.NewBuffer(buf.Bytes())))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", ctype)
		hd.ServeHTTP(writer, req)
	}
}

func testGetApiHandler(middlewares ...Middleware) http.Handler {
	api := GoAPI(false)
	api.SetLogger(nil)
	api.AddMiddleware(middlewares...)
	api.IncludeRouter(&testRouters{}, "", false)
	return api.Handler()
}

func createFile(filePath string) (string, bytes.Buffer) {
	var err error
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// 添加表单字段
	err = w.WriteField("name", filePath)
	if err != nil {
		panic(err)
	}

	// 添加文件
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	part, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		panic(err)
	}

	// 结束multipart写入
	w.Close()
	return w.FormDataContentType(), b
}

type testRouters struct {
}

func (t *testRouters) Index(input struct {
	router Router `path:"/index" method:"get"`
}) {

}

func (t *testRouters) IndexReturn(input struct {
	router Router `path:"/index/return" method:"get"`
}) testBody {
	return testBody{
		Id:   15,
		Name: "zs",
	}
}

func (t *testRouters) Middleware(input struct {
	router Router `path:"/middleware" method:"get"`
	Token  string `header:"Token"`
	Token2 string `header:"Token2"`
}) {

}

func (t *testRouters) PostData(input struct {
	router Router `path:"/post" method:"post"`
	Auth   *testAuth
	Token  string   `header:"Token"`
	Type   string   `query:"type"`
	Body   testBody `body:"json"`
}) {
}

func (t *testRouters) PostFile(input struct {
	router Router                `path:"/post/file" method:"post"`
	File   *multipart.FileHeader `file:"file"`
	Name   string                `form:"name"`
}) {
}

type testBody struct {
	Id   int
	Name string
}

type testAuth struct {
	ProjectID string `cookie:"Projectid"`
}

func (a *testAuth) ApiKey() {
}
