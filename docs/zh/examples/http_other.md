## [<<](examples.md) 如何实现h2c或是http3
程序实现了http.Handler接口可用于扩展其他的http服务
### h2c
~~~go
package main

import (
	"net/http"

	"github.com/goodluckxu-go/goapi/v2"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	api := goapi.GoAPI(true)
	h2s := &http2.Server{}
	http.ListenAndServe(":8080", h2c.NewHandler(api.Handler(), h2s))
}
~~~
### http3
~~~go
package main

import (
	"crypto/tls"

	"github.com/goodluckxu-go/goapi/v2"
	"github.com/quic-go/quic-go/http3"
)

func main() {
	api := goapi.GoAPI(true)
	h3 := &http3.Server{
		Handler: api.Handler(),
		Addr:    ":8089",
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2"},
		},
	}
	h3.ListenAndServe()
}
~~~