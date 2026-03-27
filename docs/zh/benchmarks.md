## [<<](index.md) 基准测试
GoAPI路由仿照gin实现
~~~shell
$ go test -bench="." --benchmem         
goos: windows
goarch: amd64
pkg: github.com/goodluckxu-go/goapi/v2
cpu: Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz
BenchmarkOneRouter-12                    4029139               291.4 ns/op             8 B/op          1 allocs/op
BenchmarkOneReturnRouter-12              1265724               925.9 ns/op           236 B/op          5 allocs/op
BenchmarkSecurityHTTPBearer-12           2106031               558.5 ns/op            48 B/op          3 allocs/op
BenchmarkSecurityHTTPBasic-12            2157638               545.6 ns/op            48 B/op          4 allocs/op
BenchmarkSecurityApiKey-12               2041035               579.5 ns/op            64 B/op          5 allocs/op
BenchmarkMiddlewareRouter-12             3685884               322.6 ns/op             8 B/op          1 allocs/op
BenchmarkParamRouter-12                   445142              2640 ns/op            1409 B/op         30 allocs/op
BenchmarkPostDataRouter-12                870865              1375 ns/op             502 B/op         15 allocs/op
BenchmarkPostFileRouter-12               1751235               665.8 ns/op           128 B/op          6 allocs/op
PASS
ok      github.com/goodluckxu-go/goapi/v2       14.736s
~~~