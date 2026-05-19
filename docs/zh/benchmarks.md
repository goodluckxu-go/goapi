## [<<](index.md) 基准测试
- GoAPI路由仿照gin实现
- IncludeRouter引入函数会变成0分配，引入结构体会变成1分配
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouterByIncludeFunc-12               4763630               251.5 ns/op             0 B/op          0 allocs/op
BenchmarkOneRouter-12                            4058391               291.0 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12                      1526019               773.6 ns/op           214 B/op          4 allocs/op
BenchmarkMiddlewareRouter-12                     3737335               320.6 ns/op             8 B/op          1 allocs/op
BenchmarkSecurityHTTPBearer-12                   2310204               516.5 ns/op            16 B/op          2 allocs/op
BenchmarkSecurityHTTPBasic-12                    2063950               582.3 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey-12                       2457159               495.3 ns/op            32 B/op          3 allocs/op
BenchmarkParamPath-12                            2830246               432.0 ns/op            24 B/op          2 allocs/op
BenchmarkParamPathAll-12                         2975418               393.7 ns/op            24 B/op          2 allocs/op
BenchmarkParamQuery-12                           1682284               717.3 ns/op           440 B/op          5 allocs/op
BenchmarkParamHeader-12                          2843715               413.4 ns/op            24 B/op          2 allocs/op
BenchmarkParamCookieTypeString-12                2128292               561.5 ns/op           224 B/op          4 allocs/op
BenchmarkParamCookieTypeHttpCookie-12            2131548               563.7 ns/op           216 B/op          4 allocs/op
BenchmarkPostDataRouter-12                        607201              1899 ns/op            1040 B/op         12 allocs/op
BenchmarkPostFileRouter-12                       2117958               564.6 ns/op            96 B/op          4 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       24.967s
~~~