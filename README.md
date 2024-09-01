[![GoDoc](https://godoc.org/github.com/asjard/asjard?status.svg)](https://godoc.org/github.com/asjard/asjard)
[![Go](https://github.com/asjard/asjard/actions/workflows/go.yml/badge.svg)](https://github.com/asjard/asjard/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/asjard/asjard)](https://goreportcard.com/report/github.com/asjard/asjard)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/asjard/asjard)](https://github.com/asjard/asjard/blob/main/go.mod)

## Asjard

Asjard是一个用[Go](https://go.dev/)语言实现的由[protobuf](https://protobuf.dev/)和配置驱动的微服务框架

## 安装

```bash
go get github.com/asjard/asjard
```

protobuf编译命令安装

```bash
# rest 代码生成命令
go install github.com/asjard/asjard/cmd/protoc-gen-go-rest
# rest -> grpc gateway代码生成命令
go install github.com/asjard/asjard/cmd/protoc-gen-go-rest2grpc-gw
```

## 概览

## 快速开始

- 参考[文档](https://asjard.gitbook.io/docs)
- 示例请参考[asjard-example](https://github.com/asjard/examples)

## Benchmark

![latency](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark_latency.png)
![benchmark](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark.png)
![alloc](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark_alloc.png)

[测试代码](https://github.com/asjard/benchmark)

## 三方库

下面是一些本框架中用到的开源库

| 库                                                                  | 描述               |
| ------------------------------------------------------------------- | ------------------ |
| [fasthttp](https://github.com/valyala/fasthttp)                     | http协议           |
| [fasthttp-router](https://github.com/fasthttp/router)               | http路由管理       |
| [grpc](https://google.golang.org/grpc)                              | grpc协议           |
| [protobuf](https://google.golang.org/protobuf)                      | protobuf协议       |
| [hystrix-go](https://github.com/afex/hystrix-go)                    | 熔断/降级          |
| [fsnotify](https://github.com/fsnotify/fsnotify)                    | 配置文件监听       |
| [prometheus-client-go](https://github.com/prometheus/client_golang) | prometheus监控上报 |
| [etcd](https://go.etcd.io/etcd/client/v3)                           | etcd连接           |
| [gorm](https://gorm.io/gorm)                                        | 数据库连接         |
| [redis](https://github.com/redis/go-redis/v9)                       | redis连接          |
| [yaml-v2](https://gopkg.in/yaml.v2)                                 | yaml解析           |
| [fressache](https://github.com/coocood/freecache)                   | 本地缓存           |
| [gnostic](https://github.com/google/gnostic)                        | openapiv3文档生成  |
| [cast](https://github.com/spf13/cast)                               | 配置类型转换       |
| [lumberjack](gopkg.in/natefinch/lumberjack.v2)                      | 日志防爆           |

## License

[MIT](https://github.com/asjard/asjard?tab=MIT-1-ov-file)
