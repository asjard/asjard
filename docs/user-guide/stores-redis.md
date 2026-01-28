## 配置

```yaml
asjard:
  ## 数据相关配置
  stores:
    ## redis相关配置
    redis:
      clients:
        default:
          ## host:port address
          ## address,username,password受cipherName保护
          # address: 127.0.0.1
          # username: ""
          # password: ""
          ## 加解密组件名称
          cipherName: ""
          cipherParams: {}
          options:
            ## 继承asjard.stores.redis.options
      options:
        ## ClientName will execute the `CLIENT SETNAME ClientName` command for each conn.
        # clientName: ""

        ## Protocol 2 or 3. Use the version to negotiate RESP version with redis-server.
        ## Default is 3.
        # protocol: 3

        ## Maximum number of retries before giving up.
        ## Default is 3 retries; -1 (not 0) disables retries.
        # maxRetries: 3

        ## Minimum backoff between each retry.
        ## Default is 8 milliseconds; -1 disables backoff.
        # minRetryBackoff: 8ms

        ## Maximum backoff between each retry.
        ## Default is 512 milliseconds; -1 disables backoff.
        # maxRetryBackoff: 512ms

        ## Dial timeout for establishing new connections.
        ## Default is 5 seconds.
        # dialTimeout:5s

        ## Timeout for socket reads. If reached, commands will fail
        ## with a timeout instead of blocking. Supported values:
        ##   - `0` - default timeout (3 seconds).
        ##   - `-1` - no timeout (block indefinitely).
        ##   - `-2` - disables SetReadDeadline calls completely.
        # readTimeout: 0

        ## Timeout for socket writes. If reached, commands will fail
        ## with a timeout instead of blocking.  Supported values:
        ##   - `0` - default timeout (3 seconds).
        ##   - `-1` - no timeout (block indefinitely).
        ##   - `-2` - disables SetWriteDeadline calls completely.
        # writeTimeout: 0

        ## ContextTimeoutEnabled controls whether the client respects context timeouts and deadlines.
        ## See https://redis.uptrace.dev/guide/go-redis-debugging.html#timeouts
        # contextTimeoutEnabled: false

        ## Type of connection pool.
        ## true for FIFO pool, false for LIFO pool.
        ## Note that FIFO has slightly higher overhead compared to LIFO,
        ## but it helps closing idle connections faster reducing the pool size.
        # poolFIFO: false

        ## Base number of socket connections.
        ## Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
        ## If there is not enough connections in the pool, new connections will be allocated in excess of PoolSize,
        ## you can limit it through MaxActiveConns
        # poolSize: 10

        ## Amount of time client waits for connection if all connections
        ## are busy before returning an error.
        ## Default is ReadTimeout + 1 second.
        # poolTimeout: 0s

        ## Minimum number of idle connections which is useful when establishing
        ## new connection is slow.
        ## Default is 0. the idle connections are not closed by default.
        # minIdleConns: 0

        ## Maximum number of idle connections.
        ## Default is 0. the idle connections are not closed by default.
        # maxIdleConns:0

        ## Maximum number of connections allocated by the pool at a given time.
        ## When zero, there is no limit on the number of connections in the pool.
        # maxActiveConns: 0

        ## ConnMaxIdleTime is the maximum amount of time a connection may be idle.
        ## Should be less than server's timeout.
        ##
        ## Expired connections may be closed lazily before reuse.
        ## If d <= 0, connections are not closed due to a connection's idle time.
        ##
        ## Default is 30 minutes. -1 disables idle timeout check.
        # connMaxIdleTime: 30m

        ## ConnMaxLifetime is the maximum amount of time a connection may be reused.
        ##
        ## Expired connections may be closed lazily before reuse.
        ## If <= 0, connections are not closed due to a connection's age.
        ##
        ## Default is to not close idle connections.
        # connMaxLifetime: 0s

        ## Disable set-lib on connect. Default is false.
        # disableIndentity: false

        ## Add suffix to client name. Default is empty.
        # identitySuffix: ""
```

## 使用

```go
import "github.com/asjard/asjard/pkg/stores/xredis"

// 使用默认客户端
client, err := xredis.Client()
if err != nil {
	return err
}

// 自定义客户端
// 前提是需要配置asjard.stores.redis.clients.xxx
client, err := xredis.Client(xetcd.WithClientName("xxx"))
```
