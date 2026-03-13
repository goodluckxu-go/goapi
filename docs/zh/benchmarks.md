## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                    3188806               372.2 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12              1212147              1053 ns/op             242 B/op          5 allocs/op
BenchmarkMiddlewareRouter-12             2786973               409.0 ns/op             8 B/op          1 allocs/op
BenchmarkParamRouter-12                   446024              2704 ns/op            1409 B/op         30 allocs/op
BenchmarkPostDataRouter-12                986030              1229 ns/op             537 B/op         11 allocs/op
BenchmarkPostFileRouter-12               1566078               752.9 ns/op           128 B/op          6 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       9.904s
~~~