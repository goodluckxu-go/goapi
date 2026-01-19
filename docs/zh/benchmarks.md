## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                    2944179               421.6 ns/op            40 B/op          3 allocs/op
BenchmarkOneReturnRouter-12              1000000              1291 ns/op             283 B/op          7 allocs/op
BenchmarkMiddlewareRouter-12             2554915               481.9 ns/op            56 B/op          3 allocs/op
BenchmarkParamRouter-12                   372080              3107 ns/op            1409 B/op         32 allocs/op
BenchmarkPostDataRouter-12               1000000              1347 ns/op             550 B/op         12 allocs/op
BenchmarkPostFileRouter-12               1455564               881.5 ns/op           160 B/op          8 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2  10.291s
~~~