## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                    4048297               297.2 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12              1302582               921.4 ns/op           232 B/op          5 allocs/op
BenchmarkSecurityHTTPBearer-12           2319690               522.8 ns/op            16 B/op          2 allocs/op
BenchmarkSecurityHTTPBasic-12            2134508               566.8 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey-12               2046236               585.8 ns/op            64 B/op          5 allocs/op
BenchmarkMiddlewareRouter-12             3537114               329.9 ns/op             8 B/op          1 allocs/op
BenchmarkParamRouter-12                   454556              2644 ns/op            1409 B/op         30 allocs/op
BenchmarkPostDataRouter-12                639508              1916 ns/op            1040 B/op         12 allocs/op
BenchmarkPostFileRouter-12               1754455               673.7 ns/op           128 B/op          6 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       14.862s
~~~