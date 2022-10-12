# 新的kratos工具链

kratos工具已经全部升级到v0.7 且与旧版不兼容

**如果你使用v0.7版进行开发 请务必删除gopath/bin下面所有kratos有关的二进制文件 并重新安装**

**红色框标记的 请务必删除重新安装**

![kratos-tool-list](kratos-tool-list.png)

## 安装方式

由于kratos v0.7依赖gogo的包 因此需要先安装gogo的依赖

```shell
GO111MODULE=on
```

```shell
go get -u github.com/gogo/protobuf
go get -u github.com/gogo/protobuf/protoc-gen-gogofast
go get -u github.com/gogo/googleapis
```

```shell
go get -u gitlab.com/firerocksg/xy3-kratos/tool/kratos
go get -u gitlab.com/firerocksg/xy3-kratos/tool/protobuf/protoc-gen-bm
go get -u gitlab.com/firerocksg/xy3-kratos/tool/protobuf/protoc-gen-bswagger
go get -u gitlab.com/firerocksg/xy3-kratos/tool/protobuf/protoc-gen-ecode
go get -u gitlab.com/firerocksg/xy3-kratos/tool/protobuf/protoc-gen-comet
go get -u gitlab.com/firerocksg/xy3-kratos/tool/protobuf/protoc-gen-tcp
go get -u gitlab.com/firerocksg/xy3-kratos/tool/protobuf/protoc-gen-tcp-loop
```

```shell
kratos tool install all
```

```shell
kratos new kratos-demo
```