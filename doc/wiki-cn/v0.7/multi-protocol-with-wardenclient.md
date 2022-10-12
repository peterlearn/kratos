# 多个服务发现协议共存
在新老版本过度的时期 难免会出现 服务A 调用了服务B C两个服务

服务B是新版本的服务 而且跑在k8s上 服务C是旧服务 还运行在docker和discovery上

但是服务A需要使用这两个服务 所以就有了需要两个服务发现解释器的问题

由于重写了整个服务发现逻辑 每个解释器是根据协议独立运行的

因此你可以在 v0.7版本使用多个服务发现解释器

# demo

服务C client.go
```go
package api
import (
	"context"
	"fmt"

	"gitlab.com/firerocksg/xy3-kratos/pkg/net/rpc/warden"

	"google.golang.org/grpc"
)

// AppID .
const AppID = "hello1"

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (DemoClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	if err != nil {
		fmt.Printf("ERROR!: %s init faild",AppID)
		return nil, err
	}
	return NewDemoClient(cc), nil
}

// 生成 gRPC 代码
//go:generate kratos tool protoc --grpc --bm api.proto

```


服务B client.go
```go
package api

import (
	"context"
	"gitlab.com/firerocksg/xy3-kratos/pkg/net/rpc/warden"

	"google.golang.org/grpc"
)

// AppID .
const AppID = "hello-rpc"

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (DemoClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), "kubernetes://default/hello-rpc:rpc")
	if err != nil {
		return nil, err
	}
	return NewDemoClient(cc), nil
}

// 生成 gRPC 代码
//go:generate kratos tool protoc --grpc api.proto

```

服务A main.go
```go
package main

import (
	"flag"
	
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/firerocksg/xy3-kratos/pkg/naming/discovery"
	"gitlab.com/firerocksg/xy3-kratos/pkg/naming/kubernetes"
	"gitlab.com/firerocksg/xy3-kratos/pkg/net/rpc/warden/resolver"
	"gitlab.com/firerocksg/xy3-kratos/pkg/conf/paladin"
	"gitlab.com/firerocksg/xy3-kratos/pkg/log"
	"hellohttp/internal/di"
)

func main() {
	flag.Parse()
	log.Init(&log.Config{Stdout: true}) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("hellohttp start")
	paladin.Init()


	resolver.Register(discovery.Builder())
	resolver.Register(kubernetes.Builder())

	_, closeFunc, err := di.InitApp()
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeFunc()
			log.Info("hellohttp exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

```

可以看到 上述的代码里面

service B使用k8s的服务发现 服务C使用discovery的服务发现

service A只需要在启动的时候 注册两个服务发现解释器即可

resolver.Register(discovery.Builder())

resolver.Register(kubernetes.Builder())