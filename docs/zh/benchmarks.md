## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                            3953377               319.2 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12                      1000000              1034 ns/op             251 B/op          5 allocs/op
BenchmarkMiddlewareRouter-12                     3593642               350.7 ns/op             8 B/op          1 allocs/op
BenchmarkSecurityHTTPBearer-12                   2140860               571.4 ns/op            16 B/op          2 allocs/op
BenchmarkSecurityHTTPBasic-12                    1897305               618.2 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey-12                       1992670               605.7 ns/op            64 B/op          5 allocs/op
BenchmarkParamPath-12                            2341654               535.3 ns/op            56 B/op          4 allocs/op
BenchmarkParamPathAll-12                         2347699               515.1 ns/op            56 B/op          4 allocs/op
BenchmarkParamQuery-12                           1486473               785.9 ns/op           456 B/op          6 allocs/op
BenchmarkParamHeader-12                          2023042               812.0 ns/op            56 B/op          4 allocs/op
BenchmarkParamCookieTypeString-12                1538896               747.1 ns/op           256 B/op          6 allocs/op
BenchmarkParamCookieTypeHttpCookie-12            1990146               602.3 ns/op           216 B/op          4 allocs/op
BenchmarkPostDataRouter-12                        592381              2143 ns/op            1040 B/op         12 allocs/op
BenchmarkPostFileRouter-12                       1733875               680.2 ns/op           128 B/op          6 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       25.458s
~~~