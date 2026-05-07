## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                            4051909               289.7 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12                      1566955               766.7 ns/op           210 B/op          4 allocs/op
BenchmarkMiddlewareRouter-12                     3755094               319.8 ns/op             8 B/op          1 allocs/op
BenchmarkSecurityHTTPBearer-12                   2296016               525.8 ns/op            16 B/op          2 allocs/op
BenchmarkSecurityHTTPBasic-12                    2143868               562.2 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey-12                       2479798               479.5 ns/op            32 B/op          3 allocs/op
BenchmarkParamPath-12                            2899624               411.6 ns/op            24 B/op          2 allocs/op
BenchmarkParamPathAll-12                         2964649               399.1 ns/op            24 B/op          2 allocs/op
BenchmarkParamQuery-12                           1769220               683.7 ns/op           440 B/op          5 allocs/op
BenchmarkParamHeader-12                          2834642               444.0 ns/op            24 B/op          2 allocs/op
BenchmarkParamCookieTypeString-12                2003418               575.0 ns/op           224 B/op          4 allocs/op
BenchmarkParamCookieTypeHttpCookie-12            2013277               616.3 ns/op           216 B/op          4 allocs/op
BenchmarkPostDataRouter-12                        646892              1873 ns/op            1040 B/op         12 allocs/op
BenchmarkPostFileRouter-12                       1984654               609.3 ns/op            96 B/op          4 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       24.891s
~~~