> 通过独立的 MCP 协议对外暴露工具，监听能力复用 xhttp，传输为 Streamable HTTP

## 配置

> 除[公共配置](./server.md#公共配置)外，MCP 复用 xhttp 的 CORS、超时等 HTTP 配置

```yaml
asjard:
  servers:
    mcp:
      enabled: true
      ## Streamable HTTP 路径，默认 /mcp
      # path: /mcp
      cors:
        # allowOrigins: "*"
        # allowMethods:
        #   - GET
        #   - HEAD
        #   - PUT
        #   - PATCH
        #   - POST
        #   - DELETE
        # allowHeaders:
        #   - Origin
        #   - Content-Length
        #   - Content-Type
        #   - Mcp-Session-Id
        #   - Last-Event-ID
        #   - MCP-Protocol-Version
      addresses:
        listen: 127.0.0.1:7050
        # advertise: 127.0.0.1:7050
      options:
        # readTimeout: !!str 0s
        # readHeaderTimeout: !!str 0s
        # writeTimeout: !!str 0s
        # idleTimeout: !!str 0s
        # maxHeaderBytes: !!int 0
        # maxRequestBodySize: !!int 20971520
        # disableKeepalive: !!bool false
```

## 使用

- **步骤一:** 空白导入以注册 MCP 服务

```go
import (
	_ "github.com/asjard/asjard/pkg/server/mcp"
)
```

- **步骤二:** 实现 `mcp.Handler`，通过 `McpServiceDesc` 声明工具

```go
package apiv1

import (
	"context"
	"encoding/json"

	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type greetServer interface {
	Greet(ctx context.Context, name string) (string, error)
}

type GreetAPI struct{}

func NewGreetAPI() *GreetAPI {
	return &GreetAPI{}
}

func (GreetAPI) Greet(_ context.Context, name string) (string, error) {
	if name == "" {
		name = "world"
	}
	return "hi " + name, nil
}

func (*GreetAPI) McpServiceDesc() *mcp.ServiceDesc {
	return &mcp.ServiceDesc{
		ServiceName: "api.v1.Greet",
		Name:        "Greet",
		HandlerType: (*greetServer)(nil),
		Tools: []mcp.ToolDesc{
			{
				Name:        "greet",
				Description: "say hi",
				InputSchema: json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`),
				Handler: func(ctx context.Context, req *mcpsdk.CallToolRequest, srv any, interceptor server.UnaryServerInterceptor) (*mcpsdk.CallToolResult, error) {
					var in struct {
						Name string `json:"name"`
					}
					if len(req.Params.Arguments) > 0 {
						if err := json.Unmarshal(req.Params.Arguments, &in); err != nil {
							return nil, err
						}
					}
					info := &server.UnaryServerInfo{
						Server:     srv,
						FullMethod: "/api.v1.Greet/Greet",
						Protocol:   mcp.Protocol,
					}
					handler := func(ctx context.Context, req any) (any, error) {
						return srv.(greetServer).Greet(ctx, in.Name)
					}
					if interceptor == nil {
						text, err := srv.(greetServer).Greet(ctx, in.Name)
						if err != nil {
							return nil, err
						}
						return &mcpsdk.CallToolResult{
							Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: text}},
						}, nil
					}
					resp, err := interceptor(ctx, in, info, handler)
					if err != nil {
						return nil, err
					}
					return &mcpsdk.CallToolResult{
						Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: resp.(string)}},
					}, nil
				},
			},
		},
	}
}
```

- **步骤三:** 按 `mcp` 协议注册 handler

```go
if err := server.AddHandler(apiv1.NewGreetAPI(), mcp.Protocol); err != nil {
	panic(err)
}
```

也可以用 `protoc-gen-go-mcp` 从带 `option (asjard.api.mcp_tool) = {}` 的方法生成 `McpServiceDesc`；网关再用 `protoc-gen-go-mcp2grpc` 转 gRPC：

```proto
import "github.com/asjard/protobuf/mcp.proto";

rpc SayHello(HelloRequest) returns (HelloReply) {
    option (asjard.api.mcp_tool) = {};
}
```

```go
if err := server.AddHandler(sample.NewSampleMcpAPI(), mcp.Protocol); err != nil {
	panic(err)
}
```

## 访问

启用后，MCP 客户端通过 Streamable HTTP 连接：

```
http://{asjard.servers.mcp.addresses.listen}{path}
```

默认地址为 `http://127.0.0.1:7050/mcp`。

`InputSchema` 为 JSON Schema。未设置时默认为 `{"type":"object"}`。工具调用会走该协议上的服务端拦截器链，`UnaryServerInfo.Protocol` 为 `mcp`。
