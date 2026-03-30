## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                    4015056               294.1 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12              1295706               911.4 ns/op           232 B/op          5 allocs/op
BenchmarkSecurityHTTPBearer-12           1986900               567.8 ns/op            48 B/op          3 allocs/op
BenchmarkSecurityHTTPBasic-12            2201628               536.4 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey-12               2028595               587.3 ns/op            64 B/op          5 allocs/op
BenchmarkMiddlewareRouter-12             3681998               319.6 ns/op             8 B/op          1 allocs/op
BenchmarkParamRouter-12                   443187              2660 ns/op            1409 B/op         30 allocs/op
BenchmarkPostDataRouter-12                614077              1878 ns/op            1040 B/op         12 allocs/op
BenchmarkPostFileRouter-12               1731454               681.9 ns/op           128 B/op          6 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       14.702s
~~~