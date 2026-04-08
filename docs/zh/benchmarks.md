## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                            4015159               285.3 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12                      1276286               938.9 ns/op           235 B/op          5 allocs/op
BenchmarkMiddlewareRouter-12                     3604064               319.0 ns/op             8 B/op          1 allocs/op
BenchmarkSecurityHTTPBearer-12                   2375188               504.8 ns/op            16 B/op          2 allocs/op
BenchmarkSecurityHTTPBasic-12                    2038748               587.5 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey-12                       2499870               483.7 ns/op            32 B/op          3 allocs/op
BenchmarkParamPath-12                            2930640               406.4 ns/op            24 B/op          2 allocs/op
BenchmarkParamPathAll-12                         3007446               394.5 ns/op            24 B/op          2 allocs/op
BenchmarkParamQuery-12                           1738960               685.9 ns/op           440 B/op          5 allocs/op
BenchmarkParamHeader-12                          2765163               428.7 ns/op            24 B/op          2 allocs/op
BenchmarkParamCookieTypeString-12                2102618               558.8 ns/op           224 B/op          4 allocs/op
BenchmarkParamCookieTypeHttpCookie-12            2113659               566.1 ns/op           216 B/op          4 allocs/op
BenchmarkPostDataRouter-12                        627372              1900 ns/op            1040 B/op         12 allocs/op
BenchmarkPostFileRouter-12                       2080704               575.0 ns/op            96 B/op          4 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       23.582s
~~~