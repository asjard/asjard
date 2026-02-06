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

- define protobuf

```proto
syntax = "proto3";

package api.v1.example.docs;

// The target Go package path for generated code.
option go_package = "protos-repo/example/api/v1/sample";

import "github.com/asjard/protobuf/http.proto";
import "github.com/asjard/protobuf/validate.proto";

// Sample service provides basic greeting operations.
service Sample {
    // SayHello returns a greeting message based on the provided name.
    // It supports multiple HTTP GET entrypoints for compatibility and routing flexibility.
    rpc SayHello(HelloRequest) returns (HelloReply) {
        // Dynamic path mapping (e.g., /helloworld/john)
        option (asjard.api.http) = {
            get : "/helloworld/{name}"
        };
        // Static path mapping for general greetings
        option (asjard.api.http) = {
            get : '/hello'
        };
    }
}

// HelloRequest defines the input payload for the SayHello method.
message HelloRequest {
    // The name of the person to greet.
    // Validation: Must be provided (required) and no longer than 20 characters.
    string name = 1 [ (asjard.api.validate).rules = "required,max=20" ];
}

// HelloReply defines the output payload containing the greeting result.
message HelloReply {
    // The formatted greeting string (e.g., "Hello, name!").
    string message = 1;
}
```

- implement in go

```go
package main

import (
	"context"
	"log"

	pb "protos-repo/example/api/v1/sample"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

type SampleAPI struct {
	pb.UnimplementedSampleServer
}

func NewSampleAPI() *SampleAPI {
	return &SampleAPI{}
}

func (api *SampleAPI) Start() error { return nil }
func (api *SampleAPI) Stop()        {}

// GRPC server
func (api *SampleAPI) GrpcServiceDesc() *grpc.ServiceDesc { return &pb.Sample_ServiceDesc }

// HTTP server
func (api *SampleAPI) RestServiceDesc() *rest.ServiceDesc { return &pb.SampleRestServiceDesc }

func (api *SampleAPI) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "hello " + in.Name,
	}, nil
}

func main() {
	server := asjard.New()

	if err := server.AddHandler(SampleAPI{}, rest.Protocol, grpc.Protocol); err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Start())
}
```

### See more examples and documents

- Reference [Document](https://asjard.gitbook.io/docs)
- Examples [asjard-example](https://github.com/asjard/asjard/tree/develop/_examples)
- Study in [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/asjard/asjard)

## Benchmark

- Latency

![benchmark_latency](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark_latency.png)
![concurrency_latency](https://raw.githubusercontent.com/asjard/benchmark/main/concurrency_latency.png)

- Concurrency

![concurrency](https://raw.githubusercontent.com/asjard/benchmark/main/concurrency.png)
![benchmark](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark.png)

- Allocations

![benchmark_alloc](https://raw.githubusercontent.com/asjard/benchmark/main/benchmark_alloc.png)
![concurrency_alloc](https://raw.githubusercontent.com/asjard/benchmark/main/concurrency_alloc.png)

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
