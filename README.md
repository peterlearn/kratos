![kratos](doc/img/kratos3.png)
```
COPY FROM BILIBILI
```

[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/gitlab.com/firerocksg/xy3-kratos?status.svg)](https://godoc.org/gitlab.com/firerocksg/xy3-kratos)

# Kratos

Kratos是[bilibili](https://www.bilibili.com)开源的一套Go微服务框架，包含大量微服务相关框架及工具。

> 名字来源于:《战神》游戏以希腊神话为背景，讲述由凡人成为战神的奎托斯（Kratos）成为战神并展开弑神屠杀的冒险历程。

## Features

* HTTP Blademaster：核心基于[gin](https://github.com/gin-gonic/gin)进行模块化设计，简单易用、核心足够轻量；
* GRPC Warden：基于官方gRPC开发，集成[discovery](https://github.com/bilibili/discovery)服务发现，并融合P2C负载均衡；
* Cache：优雅的接口化设计，非常方便的缓存序列化，推荐结合代理模式[overlord](https://github.com/bilibili/overlord)；
* Database：集成MySQL/HBase/TiDB，添加熔断保护和统计支持，可快速发现数据层压力；
* Config：方便易用的[paladin sdk](/docs/config/config)，可配合远程配置中心，实现配置版本管理和更新；
* Log：类似[zap](https://github.com/uber-go/zap)的field实现高性能日志库，并结合log-agent实现远程日志管理；
* Trace：基于opentracing，集成了全链路trace支持（gRPC/HTTP/MySQL/Redis/Memcached）；
* Kratos Tool：工具链，可快速生成标准项目，或者通过Protobuf生成代码，非常便捷使用gRPC、HTTP、swagger文档；


## Quick start

### Requirments

- Go version>=1.13
- 设置环境变量：
- 开启go mod：`GO111MODULE=on`
- 设置包下载代理：`GOPROXY=https://goproxy.cn,direct`
- 代理忽略公司git地址：`GONOPROXY=git.huoys.com`
- 关闭校验：`GOSUMDB=off`
- 设置GOFLAGS：`GOFLAGS="-mod=mod"`

### Required
需要安装好对应的依赖环境，以及工具：

- [go](https://golang.org/dl/)
- [protoc3.12](https://github.com/protocolbuffers/protobuf)

### Installation

1.**安装工具：**

       - 通过go get安装：

         ```shell
         go get -u gitlab.com/firerocksg/xy3-kratos/tool/kratos@master

         ```

       - 安装全部工具

         ```shell
         kratos tool install all
         ```

       - 安装protobuf


2. **生成基于kratos库的脚手架工程：**

    - 一键生成服务工程项目：

      ```shell
      kratos new [servcieName]
      ```

    - 一键生成游戏工程项目：

      ```shell
      kratos new [gameName] --game
      ```

### Build & Run

```shell
 cd [servcieName]
---
 kratos run
或
 cd [servcieName]/cmd
 go build
 ./cmd -conf ../configs
---
```

# Related

- [文档](http://kratos.huoys.com/docs/intro)
