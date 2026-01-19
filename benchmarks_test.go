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
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
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
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
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
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler(func(ctx *Context) {
		ctx.Next()
	}, func(ctx *Context) {
		ctx.Next()
	})
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkParamRouter(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/param/1/zs/li?query1=1&query2=1&query2=2", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Header", "10")
	req.Header.Add("Cookie", "cookie1=125")
	req.Header.Add("Cookie", "cookie2=346")
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkPostDataRouter(b *testing.B) {
	req, err := http.NewRequest(http.MethodPost, "/post?type=125", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	buf, _ := json.Marshal(map[string]interface{}{
		"Id":   15,
		"Name": "zs",
	})
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(bytes.NewBuffer(buf))
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkPostFileRouter(b *testing.B) {
	req, err := http.NewRequest(http.MethodPost, "/post/file", nil)
	if err != nil {
		panic(err)
	}
	ctype, buf := createFile("./README.md")
	req.Header.Set("Content-Type", ctype)
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(bytes.NewBuffer(buf.Bytes()))
		hd.ServeHTTP(writer, req)
	}
}

func testGetApiHandler(middlewares ...HandleFunc) http.Handler {
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
	router Router `paths:"/index" methods:"get"`
}) {

}

func (t *testRouters) IndexReturn(input struct {
	router Router `paths:"/index/return" methods:"get"`
}) testBody {
	return testBody{
		Id:   15,
		Name: "zs",
	}
}

func (t *testRouters) Middleware(input struct {
	router Router `paths:"/middleware" methods:"get"`
}) {

}

func (t *testRouters) Param(input struct {
	router  Router       `paths:"/param/{id}/{name:*}" methods:"get"`
	Id      string       `path:"id"`
	Name    string       `path:"name"`
	Query1  string       `query:"query1"`
	Query2  []string     `query:"query2"`
	Header  string       `header:"Header"`
	Cookie1 string       `cookie:"cookie1"`
	Cookie2 *http.Cookie `cookie:"cookie2"`
}) {
}

func (t *testRouters) PostData(input struct {
	router Router `paths:"/post" methods:"post"`
	Auth   *testAuth
	Body   testBody `body:"json"`
}) {
}

func (t *testRouters) PostFile(input struct {
	router Router                `paths:"/post/file" methods:"post"`
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
