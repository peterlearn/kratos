# 新特性
v0.7重写了当前服务发现机制

面向未来增加了基于kubernetes原生API的服务发现

同时重写了discovery etcd等当前服务发现机制 兼容旧服务

独立的负载均衡模块支持K8s RR的同时 保留kratos P2C协议兼容

(k8s服务发现默认启用RR负载平衡 discovery默认启用P2C负载平衡)

# 注意
P2C负载均衡不支持GRPC v1.30以上的版本

如果你使用discovery 请不要使用GRPC v1.30以上的版本

目前框架默认保留GRPC v1.28.1版本用于兼容旧服务

使用kubernetes协议则无版本限制 你可以使用最新的GRPC特性

# 基于kubernetes的服务发现与注册
discovery的服务发现与注册 请参考[旧版文档](../warden.md) 这里不再赘述

## 服务注册
基于kubernetes的服务发现最大的特点是省去了服务注册环节 

因为服务直接由K8S进行监控并注册到ETCD 由K8S API endpoint进行实时监控

因此我们只需要关注**服务发现**部分即可 服务启动会被K8S自动注册到API里面去

## 服务发现

和discovery用法相同 只是resolver.Register换成k8s的builder即可

因此只需要改一行代码即可实现使用k8s服务发现

main.go
```go
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/firerocksg/xy3-kratos/pkg/conf/paladin"
	"gitlab.com/firerocksg/xy3-kratos/pkg/log"

	//注意这里的两个import 分别引入resolver和kubernetes包
	"gitlab.com/firerocksg/xy3-kratos/pkg/net/rpc/warden/resolver"
	"gitlab.com/firerocksg/xy3-kratos/pkg/naming/kubernetes"

	"hellohttp/internal/di"
)

func main() {
	flag.Parse()
	log.Init(&log.Config{Stdout: true}) 
	defer log.Close()
	log.Info("hellohttp start")
	paladin.Init()

	//在这里注册解释器 和discovery一样 只不过builder换成了kubernetes
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

## client.go协议
使用client.go NewClient之前 请务必注册对应的解释器协议 参考上方代码

(比如你使用discovery 请在main里面注册 resolver.Register(discovery.Builder()))

(如果你使用kubernetes 请在main里面注册 resolver.Register(kubernetes.Builder()))

新的client.go生成之后 会把当前支持的协议都写在client.Dial这里

默认使用direct进行直连

重点变化看另外两行被注释掉的代码

这里有两个协议 分别是我们非常熟悉的discovery和新的kubernetes

warden client只需要使用不同协议的地址 即可访问不同协议服务发现中的服务

如有你有多个不同协议的服务发现 请参考[多协议共存](multi-protocol-with-wardenclient.md)

```go
package api
import (
	"context"
	"fmt"

	"gitlab.com/firerocksg/xy3-kratos/pkg/net/rpc/warden"

	"google.golang.org/grpc"
)

// AppID .
const AppID = "TODO: ADD APP ID"

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (DemoClient, error) {
	client := warden.NewClient(cfg, opts...)
	//cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	cc, err := client.Dial(context.Background(), "direct://default/127.0.0.1:9000")
	//cc, err := client.Dial(context.Background(), "kubernetes://{namespace}/{servicename}:{portname}")
	if err != nil {
        fmt.Printf("ERROR!: %s init faild",AppID)
		return nil, err
	}
	return NewDemoClient(cc), nil
}

// 生成 gRPC 代码
//go:generate kratos tool protoc --grpc --bm api.proto

```