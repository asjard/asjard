## Asjard

Asjard是一个用[Go](https://go.dev/)语言实现的由[protobuf](https://protobuf.dev/)和配置驱动的微服务框架

### 特性

- [x] 多服务端/客户端协议

  - 服务端
    - [x] grpc
    - [x] http
    - [x] pprof
  - 客户端
    - [x] grpc

- [x] 多配置源,异步实时生效

  - [x] 环境变量
  - [x] 文件
  - [x] 内存
  - [x] etcd

- [x] 自动服务注册/发现

  - 发现
    - [x] 本地配置文件服务发现
    - [x] etcd
  - 注册
    - [x] etcd

- [x] 统一日志处理

  - [x] mysql慢日志
  - [x] accesslog

- [x] 统一的错误处理

- [x] 拦截器

  - 服务端

    - [x] i18n
    - [x] accessLog
    - [x] metrics
    - [ ] 限速

  - 客户端
    - [x] 熔断降级
    - [x] 循环调用拦截
    - [ ] 限速
    - [x] http转grpc

- [x] 监控

  - [x] go_collector
  - [x] process_collector
  - [x] mysql
  - [x] api_requests_total
  - [x] api_requests_latency_seconds
  - [x] api_requests_size_bytes
  - [x] api_response_size_bytes

- [x] protobuf自动生成代码

  - [x] rest route
  - [x] openapi
  - [x] gateway
  - [x] rest转grpc
  - [ ] ts

- [x] 数据库

  - [x] mysql
  - [x] etcd
  - [x] redis
  - [ ] mongo

- [x] 多级缓存

  - [x] redis缓存
  - [x] 本地缓存

## 快速开始

> 更多示例请参考[examples](./examples)

### 编写proto协议文件

```proto
syntax = "proto3";

package api.v1.hello;

option go_package = "github.com/asjard/asjard/examples/protobuf/hello;hello";

import "github.com/asjard/protobuf/http.proto";

// 服务注释
service Hello {
    // 注释
    // 后续swagger中会被使用到
    rpc Say(SayReq) returns (SayReq) {
        // rest路由配置
        // 可以有多个
        option (asjard.api.http) = {
            get : "/v1"
        };
        option (asjard.api.http) = {
            post : "/v1/region/{region_id}/project/{project_id}/user/{user_id}"
        };
    };
}
message SayReq {
    string          region_id  = 1;
    string          project_id = 2;
    int64           user_id    = 3;
    repeated string str_list   = 4;
    repeated int64  int_list   = 5;
    SayObj          obj        = 6;
    repeated SayObj objs       = 7;
}

message SayObj {
    int32  field_int = 1;
    string field_str = 2;
}
```

### 按需生成

```sh
protoc --go_out=${GOPATH}/src -I${GOPATH}/src -I. ./*.proto

# 生成grpc需要的文件
protoc --go-grpc_out=${GOPATH}/src -I${GOPATH}/src -I. ./*.proto

# 生成rest需要的文件, rest依赖grpc生成的文件
protoc --go-rest_out=${GOPATH}/src -I${GOPATH}/src -I. ./*.proto
```

### 编写服务

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	_ "github.com/asjard/asjard/pkg/client/grpc"
	mgrpc "github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Hello 同一个方法既可以当做GRPC的handler，也可以当做http的handler
type Hello struct {
	pb.UnimplementedHelloServer
	conn pb.HelloClient
}

var _ pb.HelloServer = &Hello{}

// Bootstrap grpc客户端初始化
func (c *Hello) Bootstrap() error {
	conn, err := client.NewClient(mgrpc.Protocol, "helloGrpc").Conn()
	if err != nil {
		return err
	}
	c.conn = pb.NewHelloClient(conn)
	return nil
}

// Shutdown
func (c *Hello) Shutdown() {}

// Say .
func (c *Hello) Say(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	resp, err := c.conn.Call(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// Call .
func (c *Hello) Call(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	in.RegionId = "timeout: " + config.GetString("sleep", "")
	return in, nil
}

// RestServiceDesc rest服务描述, 如果提供rest服务，则必须提供此方法
func (Hello) RestServiceDesc() *rest.ServiceDesc {
	return &pb.HelloRestServiceDesc
}

// GrpcServiceDesc grpc服务描述, 如果提供grpc服务, 则必须提供此方法
func (Hello) GrpcServiceDesc() *grpc.ServiceDesc {
	return &pb.Hello_ServiceDesc
}

func main() {
	server := asjard.New()
	// 添加rest和grpc服务
	server.AddHandler(rest.Protocol, &Hello{}, rest.Protocol, mgrpc.Protocol)
	if err := server.Start(); err != nil {
		panic(err)
	}
}

```

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
