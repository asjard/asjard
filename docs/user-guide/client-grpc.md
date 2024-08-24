## 配置

```yaml
asjard:
  ## 客户端相关配置
  ## 不实时更新
  clients:
    ## 客户端负载均衡, 不存在或者为空，则默认为localityRoundRobin
    loadbalance: "localityRoundRobin"
    ## 同servers.interceptors配置
    # interceptors: "rest2RpcContext,cycleChainInterceptor,circuitBreaker"
    ## 同servers.certFile配置
    certFile: ""
    ## grpc客户端相关配置
    grpc:
      ## grpc客户端负载均衡相关配置
      ## 一级目录下的所有配置均可在某个服务中配置
      loadbalance: ""
      ## grpc客户端拦截器
      # interceptors: ""
      ## grpc客户端证书配置, 路径同servers.certFile配置
      certFile: ""
      ## grpc客户端相关参数
      options:
        ## keepalive相关参数
        keepalive:
          ## After a duration of this time if the client doesn't see any activity it
          ## pings the server to see if the transport is still alive.
          ## If set below 10s, a minimum value of 10s will be used instead.
          ## The current default value is infinity.
          Time: 20s
          ## After having pinged for keepalive check, the client waits for a duration
          ## of Timeout and if no activity is seen even after that the connection is
          ## closed.
          ## The current default value is 20 seconds.
          Timeout: 1s
          ## If true, client sends keepalive pings even with no active RPCs. If false,
          ## when there are no active RPCs, Time and Timeout will be ignored and no
          ## keepalive pings will be sent.
          ## false by default.
          PermitWithoutStream: false

      ## 连接到instance.name为helloGrpc这个服务的grpc客户端相关配置
      helloGrpc:
        ## helloGrpc 这个服务的负载均衡
        loadbalance: ""
        ## heeloGrpc 这个服务的证书
        certFile: ""
        ## helloGrpc 这个服务的拦截器
        interceptors: ""
        ## 同clients.grpc.options相关配置
        options: {}
```

## 使用

具体你可以参考`protoc-gen-go-rest-gw`生成的[gateway代码](https://github.com/asjard/examples/blob/main/protobuf/serverpb/server_rest_gw.pb.go)

```go
type ServerAPI struct {
	UnimplementedServerServer
	client ServerClient
}

func (api *ServerAPI) Bootstrap() error {
	conn, err := client.NewClient(grpc.Protocol, config.GetString("asjard.topology.services.server.name", "server")).Conn()
	if err != nil {
		return err
	}
	api.client = NewServerClient(conn)
	return nil
}
func (api *ServerAPI) Shutdown() {
}
```
