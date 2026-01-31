> 通过暴露grpc服务对外提供grpc服务

详细示例参考[这里](https://github.com/asjard/asjard/blob/develop/_examples/svc-example/apis/api/v1/sample.go)

## 配置

> 除[公共配置](./server.md#公共配置)外新增如下配置

```yaml
asjard:
  servers:
    grpc:
      options:
        maxConnectionIdle: 5m
        maxConnectionAge: 0s
        maxConnectionAgeGrace: 0s
        time: 10s
        timeout: 1s
```

## 示例

### 编写[protobuf](./standard-protobuf.md)文件

### 实现

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
