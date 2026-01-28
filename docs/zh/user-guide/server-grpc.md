> 通过暴露grpc服务对外提供grpc服务

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

### 编写[protobuf](./protobuf.md)文件

### 实现

```go
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


func main() {
	server := asjard.New()
	//提供grpc服务
	server.AddHandler(&HelloAPI{}, grpc.Protocol)
	// 启动服务
	if err := server.Start(); err != nil {
		panic(err)
	}
}

```
