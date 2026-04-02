## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                    4053234               290.1 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12              1268811               921.9 ns/op           235 B/op          5 allocs/op
BenchmarkSecurityHTTPBearer-12           2358626               512.4 ns/op            16 B/op          2 allocs/op
BenchmarkSecurityHTTPBasic-12            2011797               588.6 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey-12               2070991               574.9 ns/op            64 B/op          5 allocs/op
BenchmarkMiddlewareRouter-12             3716920               326.5 ns/op             8 B/op          1 allocs/op
BenchmarkParamRouter-12                   503877              2406 ns/op            1265 B/op         26 allocs/op
BenchmarkPostDataRouter-12                614800              1880 ns/op            1040 B/op         12 allocs/op
BenchmarkPostFileRouter-12               1812842               661.4 ns/op           128 B/op          6 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       14.773s
~~~