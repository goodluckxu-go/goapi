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

func BenchmarkSecurityHTTPBearer(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/security/http_bearer", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer token")
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkSecurityHTTPBasic(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/security/http_basic", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Basic dGVzdDoxMjM0NTY=")
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkSecurityApiKey(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/security/api_key", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Token", "123456")
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkParamPath(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/param/path/15", nil)
	if err != nil {
		panic(err)
	}
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkParamPathAll(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/param/pathAll/img/1.png", nil)
	if err != nil {
		panic(err)
	}
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkParamQuery(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/param/query?query=1", nil)
	if err != nil {
		panic(err)
	}
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkParamHeader(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/param/header", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Header", "10")
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkParamCookieTypeString(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/param/cookie/string", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Cookie", "cookie=125")
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkParamCookieTypeHttpCookie(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/param/cookie/httpCookie", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Cookie", "cookie=125")
	writer := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	hd := testGetApiHandler()
	for i := 0; i < b.N; i++ {
		hd.ServeHTTP(writer, req)
	}
}

func BenchmarkPostDataRouter(b *testing.B) {
	req, err := http.NewRequest(http.MethodPost, "/post", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	buf, _ := json.Marshal(map[string]interface{}{
		"Id":   15,
		"Name": "zs",
	})
	req.ContentLength = int64(len(buf))
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
	api := New(false)
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

func (t *testRouters) HTTPBearer(input struct {
	router Router `paths:"/security/http_bearer" methods:"get"`
	Auth   *testHTTPBearer
}) {
}

func (t *testRouters) HTTPBasic(input struct {
	router Router `paths:"/security/http_basic" methods:"get"`
	Auth   *testHTTPBasic
}) {

}

func (t *testRouters) ApiKey(input struct {
	router Router `paths:"/security/api_key" methods:"get"`
	Auth   *testApiKey
}) {

}

func (t *testRouters) Middleware(input struct {
	router Router `paths:"/middleware" methods:"get"`
}) {

}

func (t *testRouters) ParamPath(input struct {
	router Router `paths:"/param/path/{path}" methods:"get"`
	Path   string `path:"path"`
}) {
}

func (t *testRouters) ParamPathAll(input struct {
	router Router `paths:"/param/pathAll/{path:*}" methods:"get"`
	Path   string `path:"path"`
}) {
}

func (t *testRouters) ParamQuery(input struct {
	router Router `paths:"/param/query" methods:"get"`
	Query  string `query:"query"`
}) {
}

func (t *testRouters) ParamHeader(input struct {
	router Router `paths:"/param/header" methods:"get"`
	Header string `header:"Header"`
}) {
}

func (t *testRouters) ParamCookieTypeString(input struct {
	router Router `paths:"/param/cookie/string" methods:"get"`
	Cookie string `cookie:"cookie"`
}) {
}

func (t *testRouters) ParamCookieTypeHttpCookie(input struct {
	router Router       `paths:"/param/cookie/httpCookie" methods:"get"`
	Cookie *http.Cookie `cookie:"cookie"`
}) {
}

func (t *testRouters) PostData(input struct {
	router Router   `paths:"/post" methods:"post"`
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

type testHTTPBearer struct {
}

func (t *testHTTPBearer) HTTPBearer(token string) {
}

type testHTTPBasic struct {
}

func (t *testHTTPBasic) HTTPBasic(username, password string) {
}

type testApiKey struct {
	Token string `header:"Token"`
}

func (t *testApiKey) ApiKey() {
}
