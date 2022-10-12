## v0.7教程
* [**如何升级到新版本**](doc/wiki-cn/v0.7/how-to-update.md)
* [**其他服务如何调用kratos wardern RPC**](doc/wiki-cn/v0.7/other-service-use-wardenclient.md)
* [**多个不同协议的RPC共存**](doc/wiki-cn/v0.7/multi-protocol-with-wardenclient.md)
* [**在K8S上进行开发调试**](doc/wiki-cn/v0.7/debug-with-other-service-on-k8s.md)

## v0.7新特性
* 全新的[**服务发现**](doc/wiki-cn/v0.7/service-discovery.md)机制 支持原生K8S
* 全量使用gogo protobuf特性[**兼容新旧版本**](doc/wiki-cn/v0.7/other-compatibility.md) 解决google v2API panic问题
* [**全新的工具链**](doc/wiki-cn/v0.7/new-generator-tools.md) 支持gogo特性 生成模板
* 引入[**RPC依赖注入**](doc/wiki-cn/v0.7/RPC-DI.md) 初始化健康检查