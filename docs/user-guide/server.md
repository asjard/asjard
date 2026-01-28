> 可以通过暴露已有的服务，或者自行实现相应的协议对外暴露服务

## 已实现服务

- [x] [rest](./server-rest.md)
- [x] [grpc](./server-grpc.md)
- [x] [pprof](./server-pprof.md)

## 公共配置

```yaml
## 服务欧相关配置
asjard:
  ## 多协议服务相关配置
  ## 不实时生效，修改后需重新启动
  servers:
    ## 协议无关的服务端拦截器列表,多个拦截器以英文逗号(,)分隔
    # interceptors: ""
    ## 内建配置的拦截器
    # builtInInterceptors:
    #   - ratelimiter
    #   - metrics
    #   - accessLog
    #   - restReadEntity
    #   - restResponseHeader
    #   - i18n
    #   - trace
    ## 默认处理器
    # defaultHandlers: ""
    ## 内建配置的默认处理器
    # builtInDefaultHandlers:
    #   - health
    #   - metrics
    ## 证书文件,ASJARD_CERT_DIR下的路径
    certFile: ""
    ## 私钥文件, ASJARD_CERT_DIR下的路径
    keyFile: ""
```

## 如何实现自己的服务

### 配置约定

- 配置应都放在`asjard.servers.{自定义服务名称}`该命名空间下

### 自定义服务

> 具体可参考`core/server/server.go`中的代码

- 需实现如下两个方法

```go

// Server 每个协议需要实现的内容
type Server interface {
	// 注册
	AddHandler(handler any) error
	// 服务启动
	Start(startErr chan error) error
	// 服务停止
	Stop()
	// 服务提供的协议
	Protocol() string
	// 服务监听地址列表
	ListenAddresses() AddressConfig
	// 是否已启用
	Enabled() bool
}

// NewServerFunc 服务初始化方法
type NewServerFunc func(options *ServerOptions) (Server, error)
```

- 然后通过`AddServer`添加服务

```go
func init() {
	server.AddServer(Protocol, New)
}
```

- 具体实现可参考`pkg/server/rest`,`pkg/server/grpc`,`pkg/server/pprof`

### 使用

- **步骤一:** 编写protobuf协议文件, 详细请参考[protobuf](./protobuf.md)

```proto
// github.com/asjard/asjrd/examples/example/example.proto

syntax = "proto3";

package api.v1.hello;

option go_package = "github.com/asjard/asjard/examples/example/hellopb";

import "github.com/asjard/protobuf/http.proto";

// 需要实现的功能
service Hello {
  // 功能描述,
  // 支持markdown
  // 可渲染在openapi文档中
  rpc Say(SayReq) returns (SayResp) {
    // 如果是要对外暴露rest服务则写如下路由信息
    option (asjard.api.http) = {
      get : "/hello"
    };
    option (asjard.api.http) = {
      post : "/hello"
    };
    option (asjard.api.http) = {
      delete : "/hello/{name}"
    };
  };
}

// 请求参数
message SayReq {
  // 名称
  string name = 1;
}

// 请求返回
message SayResp { string message = 2; }

```

- **步骤二:** 编译协议文件

```bash
protoc --go_out=${GOPATH}/src -I${GOPATH}/src -I example.proto
# 如果需要暴露grpc服务则编译grpc需要的文件
protoc --go-grpc_out=${GOPATH}/src -I${GOPATH}/src -I example.proto
# 如果需要暴露rest服务则编译rest需要的文件， 依赖上一步
protoc --go-rest_out=${GOPATH}/src -I${GOPATH}/src -I example.proto
```

- **步骤三:** 编写服务

```go
// github.com/asjard/asjrd/examples/example/example.proto

package main

import (
	"context"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/examples/example/hellopb"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

// HelloAPI hello相关接口
type HelloAPI struct {
	hellopb.UnimplementedHelloServer
}

func (api *HelloAPI) Say(ctx context.Context, in *hellopb.SayReq) (*hellopb.SayResp, error) {
	return &hellopb.SayResp{
		Message: "hello " + in.Name,
	}, nil
}

// 如果需要提供grpc服务则需要实现此方法
func (api *HelloAPI) GrpcServiceDesc() *grpc.ServiceDesc {
	return &hellopb.Hello_ServiceDesc
}

// 如果需要提供rest服务则需要实现此方法
func (api *HelloAPI) RestServiceDesc() *rest.ServiceDesc {
	return &hellopb.HelloRestServiceDesc
}

func main() {
	server := asjard.New()
	// 同时提供grpc和rest服务
	server.AddHandlerV2(&HelloAPI{}, rest.Protocol, grpc.Protocol)
	// 启动服务
	if err := server.Start(); err != nil {
		panic(err)
	}
}

```

- 以上三个步骤就是编写并启动一个服务的所有流程
- 配置和访问详细信息可参考[grpc服务](server-grpc.md)和[rest服务](server-rest.md)
