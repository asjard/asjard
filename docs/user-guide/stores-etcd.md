## 配置

```yaml
asjard:
  ## 数据相关配置
  stores:
    ## etcd相关配置
    etcd:
      ## etcd列表
      clients:
        default:
          endpoints:
            - 127.0.0.1:2379
      options:
        ## AutoSyncInterval is the interval to update endpoints with its latest members.
        ## 0 disables auto-sync. By default auto-sync is disabled.
        # autoSyncInterval: 0s

        ## DialTimeout is the timeout for failing to establish a connection.
        # dialTimeout: 3s

        ## DialKeepAliveTime is the time after which client pings the server to see if
        ## transport is alive.
        # dialKeepAliveTime: 3s

        ## DialKeepAliveTimeout is the time that the client waits for a response for the
        ## keep-alive probe. If the response is not received in this time, the connection is closed.
        # dialKeepAliveTimeout: 5s

        ## MaxCallSendMsgSize is the client-side request send limit in bytes.
        ## If 0, it defaults to 2.0 MiB (2 * 1024 * 1024).
        ## Make sure that "MaxCallSendMsgSize" < server-side default send/recv limit.
        ## ("--max-request-bytes" flag to etcd or "embed.Config.MaxRequestBytes").
        # maxCallSendMsgSize: 2097152

        ## MaxCallRecvMsgSize is the client-side response receive limit.
        ## If 0, it defaults to "math.MaxInt32", because range response can
        ## easily exceed request send limits.
        ## Make sure that "MaxCallRecvMsgSize" >= server-side default send/recv limit.
        ## ("--max-request-bytes" flag to etcd or "embed.Config.MaxRequestBytes").
        # maxCallRecvMsgSize: 2147483647

        ## Username is a user name for authentication.
        # userName: ""

        ## Password is a password for authentication.
        # password: ""

        ## RejectOldCluster when set will refuse to create a client against an outdated cluster.
        # projectOldCluster: false

        ## PermitWithoutStream when set will allow client to send keepalive pings to server without any active streams(RPCs).
        # permitWithoutStream: false

        ## MaxUnaryRetries is the maximum number of retries for unary RPCs.
        # maxUnaryRetries: 0

        ## BackoffWaitBetween is the wait time before retrying an RPC.
        # backoffWaitBetween: 0s

        ## BackoffJitterFraction is the jitter fraction to randomize backoff wait time.
        # backoffJitterFraction: 0.0
        ## ASJARD_CERT_DIR路径下的文件
        # caFile: ""
        # certFile: ""
        # keyFile: ""
```

## 使用

```go
import "github.com/asjard/asjard/pkg/stores/xetcd"

// 使用默认客户端
client, err := xetcd.Client()
if err != nil {
	return err
}

// 自定义客户端
// 前提是需要配置asjard.stores.etcd.clients.xxx
client, err := xetcd.Client(xetcd.WithClientName("xxx"))
```
