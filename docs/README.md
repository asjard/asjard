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

## 快速开始

> 更多示例请参考[asjard-example](https://github.com/asjard/examples)
> 或者参考[文档](https://asjard.gitbook.io/docs)

编写[proto](https://asjard.gitbook.io/docs/protobuf/protobuf)协议文件

`server.proto`

```proto
syntax = "proto3";

package api.v1.server;

option go_package = "github.com/asjard/examples/protobuf/serverpb";

import "github.com/asjard/protobuf/http.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/wrappers.proto";

service Server {
    // 对于Server这个服务设置全局配置
    option (asjard.api.serviceHttp) = {
        group : "examples/server"
    };

    // 注释，描述这个接口的作用
    rpc Say(HelloReq) returns (HelloReq) {
        option (asjard.api.http) = {
            post : "/region/{region_id}/project/{project_id}/user/{user_id}"
        };
        option (asjard.api.http) = {
            get : "/region/{region_id}/project/{project_id}/user/{user_id}"
        };
        option (asjard.api.http) = {
            delete : "/region/{region_id}/project/{project_id}/user/{user_id}"
            writer_name : "custome_writer"
        };
    };
    // sse请求
    rpc Log(google.protobuf.Empty) returns (google.protobuf.Empty) {
        option (asjard.api.http) = {
            get : "/log"
        };
    };
    // grpc请求,不会生成rest相关代码
    rpc Call(HelloReq) returns (HelloReq) {};
}

message HelloReq {
    message Obj {
        int32  field_int = 1;
        string field_str = 2;
    }
    message Configs {
        string key_in_different_sourcer                           = 4;
    }
    message Instance {
        string              id          = 1;
        string              name        = 2;
    }
    enum Kind {
        K_A = 0;
        K_B = 1;
    }
    // 区域ID
    string region_id = 1;
    // 项目ID
    string project_id = 2;
    // 用户ID
    int64 user_id = 3;
    // 对象
    Obj obj = 6;
    // 配置
    Configs configs = 8;
    // 可选枚举参数
    optional Kind kind = 15;
    // 枚举列表
    repeated Kind kinds       = 16;
    bytes         bytes_value = 17;
    // openapi 会把这个字段解析为字符串
    uint64                     uint64_value       = 18;
}

```

按需生成

```sh
protoc --go_out=${GOPATH}/src -I${GOPATH}/src -I. ./server.proto

# 生成grpc需要的文件
protoc --go-grpc_out=${GOPATH}/src -I${GOPATH}/src -I. ./server.proto

# 生成rest需要的文件, rest依赖grpc生成的文件
protoc --go-rest_out=${GOPATH}/src -I${GOPATH}/src -I. ./server.proto

# 生成rest转grpc网关代码,依赖rest生成的文件
protoc --go-rest2grpc-gw_out=${GOPATH}/src -I${GOPATH}/src -I. ./server.proto

```

编写服务

`main.go`

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"

	// 加载etcd配置源
	_ "github.com/asjard/asjard/pkg/config/etcd"
	// 从etcd发现服务, 并把当前服务注册到etcd
	_ "github.com/asjard/asjard/pkg/registry/etcd"
	// 加载grpc服务
	"github.com/asjard/asjard/pkg/server/grpc"
	// 加载rest服务
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/examples/protobuf/serverpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ServerAPI struct {
	serverpb.UnimplementedServerServer
	exit   <-chan struct{}
	client serverpb.ServerClient
}

var _ bootstrap.Initiator = &ServerAPI{}

// Bootstrap 服务启动前会自动调用这个方法
// 当前这个方法内初始化了grpc客户端
func (api *ServerAPI) Start() error {
	conn, err := client.NewClient(grpc.Protocol, config.GetString("asjard.topology.services.examples.name", "server")).Conn()
	if err != nil {
		return err
	}
	api.client = serverpb.NewServerClient(conn)
	return nil
}

// Shutdown 服务停止会调用这里
func (api *ServerAPI) Stop() {}

// Say 接受rest请求然后去请求grpc请求
func (api *ServerAPI) Say(ctx context.Context, in *serverpb.HelloReq) (*serverpb.HelloReq, error) {
	return api.client.Call(ctx, in)
}

// Log SSE请求
func (api *ServerAPI) Log(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	rtx, ok := ctx.(*rest.Context)
	if !ok {
		return nil, status.UnsupportProtocol()
	}
	rtx.SetContentType("text/event-stream")
	rtx.SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			select {
			case <-api.exit:
				return
			default:
				w.Write([]byte(fmt.Sprintf("data: %s\n\n", time.Now())))

				if err := w.Flush(); err != nil {
					logger.Debug("client disconnected", "err", err)
					return
				}

				time.Sleep(time.Second)
			}
		}
	})
	return nil, nil
}

// Call 实时获取配置并返回
func (api *ServerAPI) Call(ctx context.Context, in *serverpb.HelloReq) (*serverpb.HelloReq, error) {
	in.Configs = &serverpb.HelloReq_Configs{
		KeyInDifferentSourcer: config.GetString("test_key", ""),
	}
	return in, nil
}

// GrpcServiceDesc 提供grpc服务,需要实现这个方法
func (api *ServerAPI) GrpcServiceDesc() *grpc.ServiceDesc {
	return &serverpb.Server_ServiceDesc
}

// RestServiceDesc 提供rest服务,需要实现这个方法
func (api *ServerAPI) RestServiceDesc() *rest.ServiceDesc {
	return &serverpb.ServerRestServiceDesc
}

func main() {
	server := asjard.New()
	// 添加grpc和rest服务
	server.AddHandler(&ServerAPI{
		exit: server.Exit(),
	}, rest.Protocol, grpc.Protocol)
	// 启动服务
	if err := server.Start(); err != nil {
		panic(err)
	}
}

```

创建配置

> 详细配置可参考[这里](https://asjard.gitbook.io/docs/pei-zhi/config)

`conf/server.yaml`

```yaml
asjard:
  ## 多协议服务相关配置
  ## 不实时生效，修改后需重新启动
  servers:
    grpc:
      enabled: true
      addresses:
        listen: 0.0.0.0:6031
    ## rest(HTTP)协议相关配置
    rest:
      enabled: true
      ## 同grpc相关配置
      addresses:
        listen: 127.0.0.1:6030
```

启动

```sh
ASJARD_CONF_DIR=${PWD}/conf go run main.go
# 或者编译后执行
go build -o example main.go && ./example

## 请求rest接口
curl 127.0.0.1:6030/api/v1/examples/server/region/region-1/project/project-1/user/1234
```

本实例内容参考[这里](https://github.com/asjard/examples/tree/main/server)

更多信息请参考[文档](https://asjard.gitbook.io/docs)

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
