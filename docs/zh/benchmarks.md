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
BenchmarkOneRouter_Func-12                               4968856               241.4 ns/op             0 B/op          0 allocs/op
BenchmarkOneReturnRouter_Func-12                         1616332               729.5 ns/op           240 B/op          3 allocs/op
BenchmarkMiddlewareRouter_Func-12                        4406822               272.5 ns/op             0 B/op          0 allocs/op
BenchmarkSecurityHTTPBearer_Func-12                      2500150               472.4 ns/op             8 B/op          1 allocs/op
BenchmarkSecurityHTTPBasic_Func-12                       2215154               541.4 ns/op            40 B/op          3 allocs/op
BenchmarkSecurityApiKey_Func-12                          2573560               461.3 ns/op            24 B/op          2 allocs/op
BenchmarkParamPath_Func-12                               3261210               362.5 ns/op            16 B/op          1 allocs/op
BenchmarkParamPathAll_Func-12                            3341767               359.5 ns/op            16 B/op          1 allocs/op
BenchmarkParamQuery_Func-12                              1885610               643.3 ns/op           432 B/op          4 allocs/op
BenchmarkParamHeader_Func-12                             3169969               377.9 ns/op            16 B/op          1 allocs/op
BenchmarkParamCookieTypeString_Func-12                   2310100               520.8 ns/op           216 B/op          3 allocs/op
BenchmarkParamCookieTypeHttpCookie_Func-12               2337934               518.6 ns/op           208 B/op          3 allocs/op
BenchmarkPostDataRouter_Func-12                           819537              1389 ns/op             336 B/op          9 allocs/op
BenchmarkPostFileRouter_Func-12                          2221296               532.3 ns/op            88 B/op          3 allocs/op
BenchmarkOneRouter_Struct-12                             3982780               296.5 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter_Struct-12                       1431386               789.0 ns/op           223 B/op          4 allocs/op
BenchmarkMiddlewareRouter_Struct-12                      3566540               333.3 ns/op             8 B/op          1 allocs/op
BenchmarkSecurityHTTPBearer_Struct-12                    2241823               535.3 ns/op            16 B/op          2 allocs/op
BenchmarkSecurityHTTPBasic_Struct-12                     1975593               597.9 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey_Struct-12                        2337379               507.2 ns/op            32 B/op          3 allocs/op
BenchmarkParamPath_Struct-12                             2847939               419.3 ns/op            24 B/op          2 allocs/op
BenchmarkParamPathAll_Struct-12                          2835387               422.8 ns/op            24 B/op          2 allocs/op
BenchmarkParamQuery_Struct-12                            1688883               712.8 ns/op           440 B/op          5 allocs/op
BenchmarkParamHeader_Struct-12                           2832990               424.5 ns/op            24 B/op          2 allocs/op
BenchmarkParamCookieTypeString_Struct-12                 2079626               723.0 ns/op           224 B/op          4 allocs/op
BenchmarkParamCookieTypeHttpCookie_Struct-12             2057552               581.3 ns/op           216 B/op          4 allocs/op
BenchmarkPostDataRouter_Struct-12                         808783              1458 ns/op             344 B/op         10 allocs/op
BenchmarkPostFileRouter_Struct-12                        2046352               584.1 ns/op            96 B/op          4 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       46.859s
~~~