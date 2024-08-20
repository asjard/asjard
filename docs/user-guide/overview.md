> 为了方便理解本框架的设计，此处对于一些概念和术语进行简要解释

## 目录结构

```sh
├── asjard.go            ## 框架入口
├── asjard_test.go
├── cmd                  ## 框架所使用的一些命令行工具存放于此目录下
├── conf                 ## 此目录下为框架所使用的配置，部分配置附带了默认值
├── core                 ## 此目录下为框架核心，均为各个组件的实现抽象(可以理解为Golang中的interface),以及很小一部分框架默认实现
├── docs                 ## 所有的文档存放于此目录下
├── go.mod               ## golang语言依赖
├── go.sum
├── pkg                  ## 对于core目录下各个组件抽象的实现
├── third_party          ## 非golang语言依赖
├── utils                ## 一些常见的工具和对于一些其他开源库库无法满足本框架需求的一些额外扩展实现
```

## 帮助文档

- [Bootstrap](bootstrap.md)
- [缓存](cache.md)
  - [本地缓存](cache-local.md)
  - [redis缓存](cache-redis.md)
- [客户端](client.md)
  - [GRPC](client-grpc.md)
  - [拦截器](client-interceptor.md)
- [动态配置](config.md)
  - [consul](config_consul.md)
  - [env](config_env.md)
  - [etcd](config_etcd.md)
  - [file](config_file.md)
- [错误](error.md)
- [初始化](initator.md)
- [日志](logger.md)
- [监控](metrics.md)
- [Protobuf协议](protobuf.md)
- [服务发现/注册](registry.md)
- [安全](security.md)
- [服务端](server.md)
  - [grpc](server-grpc.md)
  - [拦截器](server-interceptor.md)
  - [pprof](server-pprof.md)
  - [rest](server-rest.md)
- [存储](stores.md)
  - [etcd](stores-etcd.md)
  - [gorm](stores-gorm.md)
  - [redis](stores-redis.md)
  - [consul](stores-consul.md)
  - [model](stores-model.md)
