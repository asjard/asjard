> 配置驱动插件式的微服务框架,通过简单的配置即可实现相应功能或者变更程序逻辑,插件式按需加载自己想要的功能或者定制自己的插件以满足业务需求

## 背景

## 目录结构

```shell
├── CHANGELOG.md
├── Makefile
├── README.md
├── asjard.go
├── cmd
├── conf
│   ├── README.md
│   ├── cache.yaml
│   ├── certs
│   ├── cipher.yaml
│   ├── client.yaml
│   ├── config.yaml
│   ├── database.yaml
│   ├── interceptors.yaml
│   ├── logger.yaml
│   ├── registry.yaml
│   ├── servers.yaml
│   └── service.yaml
├── core ## 框架规范及默认实现
│   ├── bootstrap ## 系统启动规范
│   ├── client ## 客户端规范
│   ├── config ## 配置规范
│   ├── constant ## 常量
│   ├── logger ## 日志规范
│   ├── registry ## 服务注册发现规范
│   ├── runtime ## 运行时
│   ├── security ## 安全规范
│   └── server ## 服务规范
├── docs
│   └── doc.go
├── examples
│   ├── protobuf
│   └── server
├── go.mod
├── go.sum
├── log ## 日志目录
│   ├── asjard.log
│   └── exampleService.log ## 跟logger.filePath配置相关
├── pkg  ## 对core目录下功能的扩展实现,以及一些业务通用实现
│   ├── client
│   ├── config
│   ├── database
│   ├── logger
│   ├── registry
│   ├── security
│   ├── server
│   └── status
├── utils ## 工具包
│   ├── cast
│   ├── file.go
│   ├── ip.go
│   ├── utils.go
│   └── utils_test.go
└── version
```

## 特性

- []

## 快速开始

> 更多示例请参考[examples](./examples)

## [变更日志](CHANGELOG.md)

## 开发列表

- [x] [系统启动](./core/bootstrap/README.md)
- [x] [客户端管理](./core/client/README.md)
- [x] [配置管理](./core/config/README.md)
- [x] [日志管理](./core/logger/README.md)
- [x] [注册发现管理](./core/registry/REAME.md)
- [x] [运行时](./core/runtime/README.md)
- [x] [安全管理](./core/security/README.md)
- [x] [服务管理](./core/server/README.md)
- [x] 添加循环调用拦截器
- [ ] 熔断，限速，监控，链路追踪拦截器
- [ ] rest添加metrics接口
- [x] 所有协议添加health接口
- [ ] 添加rest服务返回自定义拦截器
- [x] server new方法使用options方式传参
- [ ] 添加测试用例，文档，cli工具
