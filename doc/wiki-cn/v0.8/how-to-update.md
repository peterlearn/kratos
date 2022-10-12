## 升级到v0.8

## 1. 升级工具链

依旧是建议大家先删除工具链的可执行文件 再重新安装新版

```shell
go get -u gitlab.com/firerocksg/xy3-kratos/tool/kratos
``` 

v0.8新增工具 protoc-gen-gin 用于生成gin的服务模板

```shell
go get -u gitlab.com/firerocksg/xy3-kratos/tool/protobuf/protoc-gen-gin

注意:
旧版使用BM代码生成方式为:
kratos tool protoc --grpc --bm api.proto

新版使用gin作为HTTP引擎 因此升级到v0.8之后 记得所有的--bm改成--gin

kratos tool protoc --grpc --gin api.proto
```

```
PS: 如果你使用go:generate
记得去你的代码里面把go:generate后面的命令里的--bm改gin

比如client.go里的

// 生成 gRPC 代码
//go:generate kratos tool protoc --grpc --bm api.proto

修改为:
//go:generate kratos tool protoc --grpc --gin api.proto
```


其他工具变化不大 但是同样建议升级到v0.8版本 可参考[**v0.7工具链升级**](../v0.7/new-generator-tools.md)

## 2. 需要修改的代码部分

### * 如果你是新的项目 直接用kratos new生成的代码模板已经完全适配新版本了

***

### * 如果是现有的项目你需要修改如下几个文件
***

## **1. 你的项目/internal/server/http/server.go**

替换所有的bm为gin 并使用gin的启动方式

具体差异请自行比对:

注意import变化 注意是cfg和 bm换成了kgin
```go
package http

import (
	"net/http"

	pb "testgin/api"
	"testgin/internal/model"
	"gitlab.com/firerocksg/xy3-kratos/pkg/conf/paladin"
	"gitlab.com/firerocksg/xy3-kratos/pkg/log"
	cfg "gitlab.com/firerocksg/xy3-kratos/pkg/net/http/config"
	kgin "gitlab.com/firerocksg/xy3-kratos/pkg/net/http/gin"
	"github.com/gin-gonic/gin"
)

var svc pb.DemoServer

// New new a bm server.
func New(s pb.DemoServer) (engine *gin.Engine, err error) {
	var (
		cfg cfg.ServerConfig
		ct paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svc = s
	engine = kgin.DefaultServer(&cfg)
	pb.RegisterDemoGinServer(engine, s)
	initRouter(engine)
	kgin.Start(engine)
	return
}

func initRouter(e *gin.Engine) {
	e.GET("/ping", ping)
	g := e.Group("/testgin")
	{
		g.GET("/start", howToStart)
	}
}

func ping(ctx *gin.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

// example for http request handler.
func howToStart(c *gin.Context) {
	k := &model.Kratos{
		Hello: "Golang 大法好 !!!",
	}
	c.JSON(200, k)
}
```

***

## **2. 你的项目internal/di/app.go**
如下:

移除http的 h.Shutdown(ctx)方法

并替换所有的bm为gin

记得改完重新wire生成一下

```go
package di

import (
	"context"
	"time"

	"testgin/internal/service"

	"gitlab.com/firerocksg/xy3-kratos/pkg/log"
	"github.com/gin-gonic/gin"
	"gitlab.com/firerocksg/xy3-kratos/pkg/net/rpc/warden"
)

//go:generate kratos tool wire
type App struct {
	svc *service.Service
	http *gin.Engine
	grpc *warden.Server
}

func NewApp(svc *service.Service, h *gin.Engine, g *warden.Server) (app *App, closeFunc func(), err error){
	app = &App{
		svc: svc,
		http: h,
		grpc: g,
	}
	closeFunc = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		if err := g.Shutdown(ctx); err != nil {
			log.Error("grpcSrv.Shutdown error(%v)", err)
		}
		//if err := h.Shutdown(ctx); err != nil {
		//	log.Error("httpSrv.Shutdown error(%v)", err)
		//}
        log.Info("httpSrv.Shutdown")
		cancel()
	}
	return
}


```

***

