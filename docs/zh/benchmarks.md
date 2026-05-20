## [<<](index.md) 基准测试
- GoAPI路由仿照gin实现
- 可以引入函数，可实现0分配
- 可以使用结构体，需要1分配
- 测试环境为 **go1.25.0 windows/amd64**
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter_Func-12                               4750810               243.8 ns/op             0 B/op          0 allocs/op
BenchmarkOneReturnRouter_Func-12                         1669477               709.4 ns/op           235 B/op          3 allocs/op
BenchmarkMiddlewareRouter_Func-12                        4384903               269.4 ns/op             0 B/op          0 allocs/op
BenchmarkSecurityHTTPBearer_Func-12                      2579572               465.1 ns/op             8 B/op          1 allocs/op
BenchmarkSecurityHTTPBasic_Func-12                       2262271               537.9 ns/op            40 B/op          3 allocs/op
BenchmarkSecurityApiKey_Func-12                          2602576               455.9 ns/op            24 B/op          2 allocs/op
BenchmarkParamPath_Func-12                               3284144               363.1 ns/op            16 B/op          1 allocs/op
BenchmarkParamPathAll_Func-12                            3329061               349.0 ns/op            16 B/op          1 allocs/op
BenchmarkParamQuery_Func-12                              1886247               633.8 ns/op           432 B/op          4 allocs/op
BenchmarkParamHeader_Func-12                             3233847               368.7 ns/op            16 B/op          1 allocs/op
BenchmarkParamCookieTypeString_Func-12                   2326702               516.5 ns/op           216 B/op          3 allocs/op
BenchmarkParamCookieTypeHttpCookie_Func-12               2307715               523.2 ns/op           208 B/op          3 allocs/op
BenchmarkPostDataRouter_Func-12                           650631              1816 ns/op            1032 B/op         11 allocs/op
BenchmarkPostFileRouter_Func-12                          2312164               517.6 ns/op            88 B/op          3 allocs/op
BenchmarkOneRouter_Struct-12                             4106521               288.7 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter_Struct-12                       1516620               771.1 ns/op           215 B/op          4 allocs/op
BenchmarkMiddlewareRouter_Struct-12                      3678409               323.2 ns/op             8 B/op          1 allocs/op
BenchmarkSecurityHTTPBearer_Struct-12                    2289836               521.8 ns/op            16 B/op          2 allocs/op
BenchmarkSecurityHTTPBasic_Struct-12                     2041200               586.8 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey_Struct-12                        2475114               488.6 ns/op            32 B/op          3 allocs/op
BenchmarkParamPath_Struct-12                             2884135               410.3 ns/op            24 B/op          2 allocs/op
BenchmarkParamPathAll_Struct-12                          2987028               403.9 ns/op            24 B/op          2 allocs/op
BenchmarkParamQuery_Struct-12                            1664571               724.0 ns/op           440 B/op          5 allocs/op
BenchmarkParamHeader_Struct-12                           2439639               432.0 ns/op            24 B/op          2 allocs/op
BenchmarkParamCookieTypeString_Struct-12                 2088534               574.0 ns/op           224 B/op          4 allocs/op
BenchmarkParamCookieTypeHttpCookie_Struct-12             2097565               579.2 ns/op           216 B/op          4 allocs/op
BenchmarkPostDataRouter_Struct-12                         618434              1886 ns/op            1040 B/op         12 allocs/op
BenchmarkPostFileRouter_Struct-12                        2095598               573.8 ns/op            96 B/op          4 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       46.308s
~~~