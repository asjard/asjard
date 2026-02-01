> 可以通过暴露已有的服务，或者自行实现相应的协议对外暴露服务

详细示例参考[这里](https://github.com/asjard/asjard/blob/develop/_examples/svc-example/apis/api/v1/user.go)

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

- **步骤一:** 编写protobuf协议文件, 详细请参考[protobuf](./standard-protobuf.md)

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
package apiv1

import (
	"context"

	pb "protos-repo/example/api/v1/sample"

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

// GRPC服务
func (api *SampleAPI) GrpcServiceDesc() *grpc.ServiceDesc { return &pb.Sample_ServiceDesc }

// HTTP服务
func (api *SampleAPI) RestServiceDesc() *rest.ServiceDesc { return &pb.SampleRestServiceDesc }

func (api *SampleAPI) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "hello " + in.Name,
	}, nil
}


```

- 以上三个步骤就是编写并启动一个服务的所有流程
- 配置和访问详细信息可参考[grpc服务](server-grpc.md)和[rest服务](server-rest.md)
