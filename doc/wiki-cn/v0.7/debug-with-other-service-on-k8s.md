# 背景
在微服务体系里面 一个或多个服务由不同的开发者分工协调

所以 我们经常会有如下的场景:

开发者A 开发A服务 需要调用B服务

开发者B 开发B服务 需要被A服务调用

此时B服务已经开发完成 并且部署到开服务器的K8s集群上去

A服务还在开发 需要一边调试 一边开发

那么此时就有一个问题 B服务在K8s上 A需要调用B 如果A也部署到k8s上是可以调用B 但是A还需要调试代码 就很不方便

那么有没有办法 可以让A服务一边调试 一边调用K8S上的B服务呢

# 在k8s上进行开发与调试
所以 引入今天的主题 如何使用kubectl的port-forward功能 实现对K8s集群上的服务进行调用调试

工具 kubectl 请自行下载 并配置开发测试服的配置文件

这里主要讲如何使用

首先 我部署hello-rpc服务到K8s上去

PS:以下使用的命令及参数不做赘述 需要了解的可自行查阅kubectl的文档 如有不懂可以开发群组里咨询

使用 ```kubectl get pods --selector "app=hello-rpc" --output=name```

输出 
```
❯ kubectl get pods --selector "app=hello-rpc" --output=name
pod/hello-rpc-7d6f8f565f-q6d6r
```

pod/hello-rpc-7d6f8f565f-q6d6r这个就是我的服务部署上去之后k8s给我生成的pod名

之后我们使用 ```kubectl port-forward``` 来把k8s上pod的端口映射到我们本地端口

这里我把pod的9000端口 映射到我本地 localhost的9000端口

```kubectl port-forward pod/hello-rpc-7d6f8f565f-q6d6r 9000:9000```

那么此时 你访问localhost:9000 或者 127.0.0.1:9000 就相当于访问了k8s里面 pod/hello-rpc-7d6f8f565f-q6d6r 的9000端口

既然已经把服务映射到本地了 那么调试的时候 

我在服务A使用的 client.go里面 修改地址为direct://default/127.0.0.1:9000 即可进行调试

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
	cc, err := client.Dial(context.Background(), "direct://default/127.0.0.1:9000")
	//cc, err := client.Dial(context.Background(), "kubernetes://default/hello-rpc:rpc")
	if err != nil {
		return nil, err
	}
	return NewDemoClient(cc), nil
}

// 生成 gRPC 代码
//go:generate kratos tool protoc --grpc api.proto

```

打好断点 启动debug 可以轻松在本地访问集群内部的服务 并返回数据进行调试


当然kubectl port-forward这条命令也可以合并成一个 更加简洁

```
kubectl port-forward $(kubectl get pods --selector "app={appname}" --output=name) host-port:pod-port
```