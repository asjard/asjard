[![GoDoc](https://godoc.org/github.com/asjard/asjard?status.svg)](https://godoc.org/github.com/asjard/asjard)
[![Go](https://github.com/asjard/asjard/actions/workflows/go.yml/badge.svg)](https://github.com/asjard/asjard/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/asjard/asjard)](https://goreportcard.com/report/github.com/asjard/asjard)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/asjard/asjard)](https://github.com/asjard/asjard/blob/main/go.mod)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/asjard/asjard)

## Asjard

Asjard是一个用[Go](https://go.dev/)语言实现的由[protobuf](https://protobuf.dev/)和配置驱动的微服务框架

## 安装

```bash
go get github.com/asjard/asjard
```

## 使用帮助

查看帮助

```bash
make help
## 或者
make
```

```bash
Commands:
  update                       更新本地代码
  build_cipher_aes             生成asjard_cipher_aes命令
  build_gen_go_rest            生成protoc-gen-go-rest命令
  build_gen_go_validate        生成protoc-gen-go-validate命令
  build_gen_go_asynq           生成protoc-gen-go-rest命令
  build_gen_go_rabbitmq        生成protoc-gen-go-rabbitmq命令
  build_gen_go_rest2grpc_gw    生成protoc-gen-go-rest2grpc-gw命令
  build_gen_ts                 生成protoc-gen-ts命令
  build_gen_ts_enum            生成protoc-gen-ts-enum命令
  build_gen_ts_umi             生成protoc-gen-ts-umi命令
  gen_proto                    生成协议文件
  github_workflows_dependices  github workflows 依赖环境
  github_workflows_test        github workflow 运行测试用例
  test                         运行测试用例
  gocyclo                      圈复杂度检测
  govet                        静态检查
  clean                        清理
  configure                    配置
  help                         使用帮助
Envs:
  BIFROST_DIR                bifrost所在目录,结尾不要有/                   默认: .         当前: ./third_party/bifrost
  PROTO_DIR                  proto协议文件目录,结尾不要有/                  默认: .         当前: .
  GEN_PROTO_GO               是否根据protobuf文件生成*.pb.go文件            默认: true      当前: true
  GEN_PROTO_GO_OUT           生成的*.pb.go文件输出目录                      默认: ./..      当前: ./..
  GEN_PROTO_GO_OPT           生成*.pb.go需要的参数                          默认:           当前:
  GEN_PROTO_GO_GRPC          是否根据protobuf文件生成*_grpc.pb.go文件       默认: true      当前: true
  GEN_PROTO_GO_GRPC_OPT      生成*_grpc.pb.go所需要的参数                   默认:           当前:
  GEN_PROTO_GO_REST          是否根据protobuf文件生成*_rest.pb.go文件       默认: true      当前: true
  GEN_PROTO_GO_REST_OPT      生成*_rest.pb.go所需要的参数                   默认:           当前:
  GEN_PROTO_GO_REST_GW       是否根据protobuf文件生成*_rest_gw.pb.go文件    默认: true      当前: true
  GEN_PROTO_GO_REST_GW_OPT   生成*_rest_gw.pb.go所需要的参数                默认:           当前:
  GEN_PROTO_TS               是否根据protobuf文件生成*.d.tsx文件            默认: false     当前: false
  GEN_PROTO_TS_OPT           生成*.d.tsx所需要的参数                        默认:           当前:
  GEN_PROTO_TS_ENUM          是否根据protobuf文件生成*.enum.tsx文件         默认: false     当前: false
  GEN_PROTO_TS_ENUM_OPT      生成*.enum.tsx所需要的参数                     默认:           当前:
  GEN_PROTO_TS_UMI           是否根据protobuf生成*.umi.tsx文件              默认: fasle     当前: fasle
  GEN_PROTO_TS_UMI_OPT       生成*.umi.tsx所需要的参数                      默认:           当前:
  GEN_PROTO_TS_OUT           生成的*.pb.ts文件输出目录                      默认: ./..      当前: ./..
  GEN_PROTO_GO_ASYNQ         是否根据protobuf文件生成*_asynq.pb.go文件      默认: false     当前: false
  GEN_PROTO_GO_ASYNQ_OPT     生成*_asynq.pb.go所需要的参数                  默认:           当前:
  GEN_PROTO_GO_RABBITMQ      是否根据protobuf文件生成*_rabbitmq.pb.go文件   默认: false     当前: false
  GEN_PROTO_GO_RABBITMQ_OPT  生成*_rabbitmq.pb.go所需要的参数               默认:           当前:
  GEN_PROTO_GO_VALIDATE      是否根据protobuf文件生成*_validate.pb.go文件   默认: true      当前: true
  GEN_PROTO_GO_VALIDATE_OPT  生成*_validate.pb.go文件的参数                 默认:           当前:
  GOOS                       运行环境,可选:linux,darwin,windows             默认: linux     当前: linux
  CGO_ENABLED                是否开启CGO,可选:0,1                           默认: 0         当前: 0
  BUILD_DIR                  编译目录                                       默认: .         当前: .
  DEBUG                      是否开启DEBUG                                  默认: false     当前: false
  PROJECT_NAME               项目名称                                       默认: bifrost   当前: bifrost
  DEPLOY_ENV                 部署环境,可选:dev,sit,uat,pro                  默认: dev       当前: dev
  SERVICE_NAME               服务名称                                       默认: bifrost   当前: bifrost
```

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
| [asynq](github.com/hibiken/asynq)                                   | 消息队列           |
| [rabbitmq](github.com/streadway/amqp)                               | 消息队列           |

## License

[MIT](https://github.com/asjard/asjard?tab=MIT-1-ov-file)
