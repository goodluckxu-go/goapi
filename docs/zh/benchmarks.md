## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                    3216332               368.3 ns/op            40 B/op          3 allocs/op
BenchmarkOneReturnRouter-12              1000000              1014 ns/op             283 B/op          7 allocs/op
BenchmarkMiddlewareRouter-12             1000000              1009 ns/op             184 B/op         10 allocs/op
BenchmarkPostDataRouter-12                507800              2333 ns/op            1170 B/op         23 allocs/op
BenchmarkPostFileRouter-12               1561086               761.6 ns/op           160 B/op          8 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi  6.587
~~~