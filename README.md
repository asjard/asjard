[![GoDoc](https://godoc.org/github.com/asjard/asjard?status.svg)](https://godoc.org/github.com/asjard/asjard)
[![Go](https://github.com/asjard/asjard/actions/workflows/go.yml/badge.svg)](https://github.com/asjard/asjard/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/asjard/asjard)](https://goreportcard.com/report/github.com/asjard/asjard)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/asjard/asjard)](https://github.com/asjard/asjard/blob/main/go.mod)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/asjard/asjard)

## Asjard

Asjard is a protobuf-driven microservice framework implemented in Go that provides a unified platform for building scalable distributed services. Orchestrates multiple protocols (REST, gRPC, message queues), dynamic configuration management, code generation pipelines, and comprehensive observability features through a single configuration-driven interface.

## Features

- Code Generation
- Dynamic Configuration
- Data Stores
- Error Management
- OpenAPI Document

## Quick Start

### Install

```bash
go get -u github.com/asjard/asjard
```

### Running in asjard

```go
package main

import (
	"github.com/asjard/asjard"
	"github.com/asjard/asjard/pkg/server/grpc"
	"your_protobuf_generated_dir/pb"
)

type ExampleAPI struct{}

func NewExampleAPI() *ExampleAPI {
	return &ExampleAPI{}
}

func(ExampleAPI) Start() error {return nil}
func (ExampleAPI) Stop() {}
func (ExampleAPI) GrpcServiceDesc() *grpc.ServiceDesc{
	return pb.Example_ServiceDesc
}

func main() {
	server := asjard.New()

	if err := server.AddHandlers(grpc.Protocol,
		apiV1.NewExampleAPI(svcCtx)); err != nil {
		log.Fatal(err)
	}
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
```

### See more examples and documents

- Reference [Document](https://asjard.gitbook.io/docs)
- Examples [asjard-example](https://github.com/asjard/asjard/tree/develop/_examples)
- Study in [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/asjard/asjard)

## Benchmark

![latency](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark_latency.png)
![benchmark](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark.png)
![alloc](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark_alloc.png)

More information see [TestCode](https://github.com/asjard/benchmark)

## ThirdParty

Here are some open source libraries used in this framework

| Repo                                                                | Description                 |
| ------------------------------------------------------------------- | --------------------------- |
| [fasthttp](https://github.com/valyala/fasthttp)                     | http protocol               |
| [fasthttp-router](https://github.com/fasthttp/router)               | http route management       |
| [grpc](https://google.golang.org/grpc)                              | grpc protocol               |
| [protobuf](https://google.golang.org/protobuf)                      | protobuf protocol           |
| [hystrix-go](https://github.com/afex/hystrix-go)                    | circuit breaker             |
| [fsnotify](https://github.com/fsnotify/fsnotify)                    | configration file listen    |
| [prometheus-client-go](https://github.com/prometheus/client_golang) | prometheus                  |
| [etcd](https://go.etcd.io/etcd/client/v3)                           | etcd client                 |
| [gorm](https://gorm.io/gorm)                                        | database client             |
| [redis](https://github.com/redis/go-redis/v9)                       | redis client                |
| [yaml-v2](https://gopkg.in/yaml.v2)                                 | yaml parser                 |
| [fressache](https://github.com/coocood/freecache)                   | local cache                 |
| [gnostic](https://github.com/google/gnostic)                        | openapiv3 document generate |
| [cast](https://github.com/spf13/cast)                               | type convert                |
| [lumberjack](gopkg.in/natefinch/lumberjack.v2)                      | log                         |
| [asynq](github.com/hibiken/asynq)                                   | asynq                       |
| [rabbitmq](github.com/streadway/amqp)                               | rabbitmq                    |

## License

[MIT](https://github.com/asjard/asjard?tab=MIT-1-ov-file)
