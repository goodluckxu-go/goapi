## [<<](index.md) 快速入门


### 要求
- Go 1.18 及以上版本
### 安装
1. 下载并安装 GoAPI：
~~~
go get -u github.com/goodluckxu-go/goapi/v2
~~~
2. 将 GoAPI 引入到代码中：
~~~go
import "github.com/goodluckxu-go/goapi/v2"
~~~
3. (可选）GoAPI还有其他包可使用：
~~~go
import (
	"github.com/goodluckxu-go/goapi/v2/lang" //语音包
	"github.com/goodluckxu-go/goapi/v2/response" // 扩展返回包
)

func main() {
	goapi.Colorful = false // 关闭默认日志控制台颜色
}
~~~
### 运行示例
1. 创建项目并且 cd 到项目目录中
~~~shell
mkdir project && cd project
~~~
2. 初始化 go mod
~~~shell
go mod init project
~~~
3. 启动项目
~~~shell
go run main.go
~~~
### 开始
首先创建一个`main.go`文件，代码如下：
~~~go
package main

import (
	"log"

	"github.com/goodluckxu-go/goapi/v2"
	"github.com/goodluckxu-go/goapi/v2/lang"
)

func main() {
	//api := goapi.New(true)
	api := goapi.Default(true)
	api.SetLang(&lang.ZhCn{})
	api.SetResponseMediaType("application/json")
	api.IncludeRouter(&Example{}, "/quick", true)
	// listen and serve on 0.0.0.0:8080
	if err := api.Run(); err != nil {
		log.Fatal(err)
	}
}

type Example struct{}

func (e *Example) Ping(input struct {
	router goapi.Router `paths:"/ping" methods:"GET" summery:"首页"`
}) ExampleResp {
	return ExampleResp{
		Msg: "pong",
    }
}

type ExampleResp struct {
	Msg string
}
~~~
然后，执行 `go run main.go` 命令来运行代码：
~~~shell
# 运行 main.go 并且在浏览器中访问 0.0.0.0:8080/ping
go run main.go
~~~
