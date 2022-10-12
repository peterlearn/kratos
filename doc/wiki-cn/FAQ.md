# 安装错误整理
### 出现 `missing go.sum entry`
    设置go env -w "GOFLAGS"="-mod=mod" 就可以了

### 如果`kratos run`无法启动，检查

    1.[serviceName]/api是否有api.xx.go文件(可在目录下执行`go generate` 生成)
    2.[serviceName]/internal/di是否有wire_gen.go文件(可在目录下执行`go generate` 生成)
    3. ...\api\client.go:15:68: undefined: DemoClient(可在目录下执行`go generate` 生成)
    4. go generate ./...   go mod tidy
### 安装失败，提示go mod 错误

执行
```shell
go get -u gitlab.com/firerocksg/xy3-kratos/tool/kratos
```
出现以下错误时
```shell
go: github.com/prometheus/client_model@v0.0.0-20190220174349-fd36f4220a90: parsing go.mod: missing module line
go: github.com/remyoudompheng/bigfft@v0.0.0-20190806203942-babf20351dd7e3ac320adedbbe5eb311aec8763c: parsing go.mod: missing module line
```
如果你使用了https://goproxy.io/ 代理,那你要使用其他代理来替换它，然后删除GOPATH目录下的mod缓存文件夹（`go clean --modcache`）,然后重新执行安装命令

代理列表

```
export GOPROXY=https://mirrors.aliyun.com/goproxy/
export GOPROXY=https://goproxy.cn/
export GOPROXY=https://goproxy.io/
```

