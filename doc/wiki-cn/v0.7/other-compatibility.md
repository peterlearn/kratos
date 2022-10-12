# 背景

当我们还在追赶k8s特性的时候 google已经悄然发布了protobuf v2API协议

变化非常大 主要引入命名空间 文件反射 和新的注册初始化机制

因此导致使用新版protobuf不能兼容我们的服务 

当两个服务proto文件名相同的时候

protobuf会认为在同一个命名空间里 导致2个服务panic

当你看到你的服务启动的时候 出现:

```
A future release will panic on registration conflicts. See: https://developers.google.com/protocol-buffers/docs/reference/go/faq#namespace-conflict
```

那就是你的protobuf版本的问题了 

google的文档说v1.3.4是最后一个v2API preview版本 实测v1.3.5也没问题

当使用1.3.6以上的版本就会出现panic

所以推荐大家在v2还没稳定下来之前 尽量避免使用v1.3.5以上的protobuf版本

# kratos v0.7 兼容性

kratos在v0.7 全量使用了gogo protobuf的特性来规避v2 protobuf带来的问题

后续gogo 更新到v2API之后 我们开发组也会跟进这部分特性

当前主要解决问题场景

* A服务使用google protobuf且版本高于1.3.5 B服务使用kratos A调用B 可实现完美兼容
* A服务使用kratos google protobuf且版本高于1.3.5 A调用B 可实现完美兼容
* A服务使用kratos v0.5及以下版本 B服务使用kratos v0.7 A调用B 可实现完美兼容
* A服务使用kratos v0.7及以上版本 B服务使用kratos v0.5及以下版本 A调用B 可实现完美兼容