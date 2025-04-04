> 文件配置，需支持yaml,json文件格式配置，其中变量名称使用驼峰式

## 配置所在目录

- 如果配置了环境变量`ASJARD_CONF_DIR`则读取该目录及子目录下的所有文件
- 否则读取环境变量`ASJARD_HOME_DIR`的值并拼接`conf`目录,读取该目录下及子目录下的所有文件
- 如果以上两个环境变量都没有设置,则读取`可执行程序`平级目录下的`conf`目录下及子目录下的所有文件

## 支持文件格式

- [x] yaml,yml
- [ ] json
- [ ] ini
- [ ] prop,properties

## 配置说明

### 服务相关配置

```yml
asjard:
  ## 项目名称
  ## 一个项目下可能会有多个服务
  ## 不实时生效，修改后需重新启动服务
  app: asjard
  ## 当前部署环境，例如: dev, sit, uat,rc,pro等
  environment: dev
  ## 部署区域,例如: east-1, east-2
  ## 表示不同地域，内网不互通，只能通过公网相互连接的不同区域
  region: default
  ## 可用区，例如: az-1,az-2
  ## 表示同一区域内，或者同一个机房内，可以内网互通
  avaliablezone: default
  ## 站点地址
  ## 例如: https://github.com/${asjard.app}/${asjard.instance.name}
  website: "https://github.com/asjard/asjard"
  ## 服务描述
  ## 可以写markdown格式内容
  ## 在openapi文档中可以渲染
  desc: |
    这里是服务描述
  ## 服务实例详情
  instance:
    ## 服务名称
    name: asjard
    ## 服务版本
    version: 1.0.0
    ## 自定义服务源数据
    metadata:
      tenantName: ${asjard.app}
```

### 协议相关配置

```yml
asjard:
  ## 多协议服务相关配置
  ## 不实时生效，修改后需重新启动
  servers:
    ## 协议无关的服务端拦截器列表,多个拦截器以英文逗号(,)分隔
    ## 默认为accessLog
    interceptors: "accessLog"
    ## grpc相依相关配置
    grpc:
      ## grpc协议相关拦截器配置，如果不配置则使用全局拦截器,多个拦截器配置方式同全局拦截器配置
      ## interceptors: "accessLog"
      ##  默认服务请求处理
      # defaultHandlers: "health"
      ## 是否启用grpc服务
      enabled: true
      ## 证书文件名称, 相对于配置所在目录的certs目录下文件路径名称
      certFile: ""
      ## 私钥文件名称, 路径同certFile配置
      keyFile: ""
      ## 监听地址
      ## 约定listen配置项为服务监听地址
      ## advertise配置项为跨区域访问地址
      addresses:
        ## 本机监听地址
        ## 如果配置为"0.0.0.0"的IPv4地址或者"::"的IPv6地址,注册到注册中的地址会被修改为
        ## 网卡中读取到的第一个IPv4地址，或者第一个IPv6地址
        listen: 127.0.0.1:7010
        ## 垮区域访问地址,可以为IP地址加端口，或者域名
        advertise: 47.121.0.5:8080
      ## 服务启动相关参数配置
      options:
        ## keepalive相关配置
        keepaliveParams:
          ## MaxConnectionIdle is a duration for the amount of time after which an
          ## idle connection would be closed by sending a GoAway. Idleness duration is
          ## defined since the most recent time the number of outstanding RPCs became
          ## zero or the connection establishment.
          ## The current default value is infinity.
          maxConnectionIdle: !!str 5m
          ## MaxConnectionAge is a duration for the maximum amount of time a
          ## connection may exist before it will be closed by sending a GoAway. A
          ## random jitter of +/-10% will be added to MaxConnectionAge to spread out
          ## connection storms.
          ## The current default value is infinity.
          maxConnectionAge: !!str 0s
          ## MaxConnectionAgeGrace is an additive period after MaxConnectionAge after
          ## which the connection will be forcibly closed.
          ## The current default value is infinity.
          maxConnectionAgeGrace: !!str 0s
          ## After a duration of this time if the server doesn't see any activity it
          ## pings the client to see if the transport is still alive.
          ## If set below 1s, a minimum value of 1s will be used instead.
          ## The current default value is 2 hours.
          time: !!str 10s
          ## After having pinged for keepalive check, the server waits for a duration
          ## of Timeout and if no activity is seen even after that the connection is
          ## closed.
          ## The current default value is 20 seconds.
          timeout: !!str 1s
    ## pprof相关配置
    pprof:
      ## 同grpc相关配置
      enabled: false
      ## 同grpc相关配置
      addresses:
        listen: 127.0.0.1:7020
    ## rest(HTTP)协议相关配置
    rest:
      ## 同grpc拦截器相关配置
      interceptors: "accessLog,restReadEntity,restResponseHeader"
      ## 同grpc相关配置
      enabled: true
      ## 同grpc相关配置
      certFile: ""
      ## 同grpc相关配置
      keyFile: ""
      ## 文档相关配置
      doc:
        ## 错误页地址, 如果错误返回，默认可以解决错误的文档地址
        ## 如果不配置则使用website配置
        errPage: ""
      ## 是否开启openapi
      openapi:
        enabled: false
        ## openapi.yml文件可以被打开的页面
        ## 默认为: https://petstore.swagger.io/?url=http://%s/openapi.yml
        page: ""
        termsOfServer: ""
        license:
          name: ""
          url: ""
      ## 跨域相关配置
      cors:
        ## 允许的源
        allowOrigins:
          - *
        allowMethods:
          - GET
          - HEAD
          - POST
          - PUT
          - PATCH
          - DELETE
        allowHeaders:
          - Origin
          - Content-Length
          - Content-Type
        allowCredentials: true
        maxAge: 12h
      ## 同grpc相关配置
      addresses:
        listen: 127.0.0.1:7030
        # advertise: 127.0.0.1:7030
      ## 服务启动相关配置
      # options:
      #   ## The maximum number of concurrent connections the server may serve.
      #   ##
      #   ## DefaultConcurrency is used if not set.
      #   ##
      #   ## Concurrency only works if you either call Serve once, or only ServeConn multiple times.
      #   ## It works with ListenAndServe as well.
      #   Concurrency: !!int 262144

      #   ## Per-connection buffer size for requests' reading.
      #   ## This also limits the maximum header size.
      #   ##
      #   ## Increase this buffer if your clients send multi-KB RequestURIs
      #   ## and/or multi-KB headers (for example, BIG cookies).
      #   ##
      #   ## Default buffer size is used if not set.
      #   ReadBufferSize: !!int 4096

      #   ## Per-connection buffer size for responses' writing.
      #   ##
      #   ## Default buffer size is used if not set.
      #   WriteBufferSize: !!int 4096

      #   ## ReadTimeout is the amount of time allowed to read
      #   ## the full request including body. The connection's read
      #   ## deadline is reset when the connection opens, or for
      #   ## keep-alive connections after the first byte has been read.
      #   ##
      #   ## By default request read timeout is unlimited.
      #   ReadTimeout: !!str 3s

      #   ## WriteTimeout is the maximum duration before timing out
      #   ## writes of the response. It is reset after the request handler
      #   ## has returned.
      #   ##
      #   WriteTimeout: !!str 1h

      #   ## IdleTimeout is the maximum amount of time to wait for the
      #   ## next request when keep-alive is enabled. If IdleTimeout
      #   ## is zero, the value of ReadTimeout is used.
      #   IdleTimeout: !!str 0s

      #   ## Maximum number of concurrent client connections allowed per IP.
      #   ##
      #   ## By default unlimited number of concurrent connections
      #   ## may be established to the server from a single IP address.
      #   MaxConnsPerIP: !!int 0

      #   ## Maximum number of requests served per connection.
      #   ##
      #   ## The server closes connection after the last request.
      #   ## 'Connection: close' header is added to the last response.
      #   ##
      #   ## By default unlimited number of requests may be served per connection.
      #   MaxRequestsPerConn: !!int 0

      #   ## MaxIdleWorkerDuration is the maximum idle time of a single worker in the underlying
      #   ## worker pool of the Server. Idle workers beyond this time will be cleared.
      #   MaxIdleWorkerDuration: !!str 10m

      #   ## Period between tcp keep-alive messages.
      #   ##
      #   ## TCP keep-alive period is determined by operation system by default.
      #   TCPKeepalivePeriod: !!str 0s

      #   ## Maximum request body size.
      #   ##
      #   ## The server rejects requests with bodies exceeding this limit.
      #   ##
      #   ## Request body size is limited by DefaultMaxRequestBodySize by default.
      #   ## 4 * 1024 * 1024
      #   MaxRequestBodySize: !!int 4194304

      #   ## Whether to disable keep-alive connections.
      #   ##
      #   ## The server will close all the incoming connections after sending
      #   ## the first response to client if this option is set to true.
      #   ##
      #   ## By default keep-alive connections are enabled.
      #   DisableKeepalive: !!bool false

      #   ## Whether to enable tcp keep-alive connections.
      #   ##
      #   ## Whether the operating system should send tcp keep-alive messages on the tcp connection.
      #   ##
      #   ## By default tcp keep-alive connections are disabled.
      #   TCPKeepalive: !!bool false

      #   ## Aggressively reduces memory usage at the cost of higher CPU usage
      #   ## if set to true.
      #   ##
      #   ## Try enabling this option only if the server consumes too much memory
      #   ## serving mostly idle keep-alive connections. This may reduce memory
      #   ## usage by more than 50%.
      #   ##
      #   ## Aggressive memory usage reduction is disabled by default.
      #   ReduceMemoryUsage: !!bool false

      #   ## Rejects all non-GET requests if set to true.
      #   ##
      #   ## This option is useful as anti-DoS protection for servers
      #   ## accepting only GET requests and HEAD requests. The request size is limited
      #   ## by ReadBufferSize if GetOnly is set.
      #   ##
      #   ## Server accepts all the requests by default.
      #   GetOnly: !!bool false

      #   ## Will not pre parse Multipart Form data if set to true.
      #   ##
      #   ## This option is useful for servers that desire to treat
      #   ## multipart form data as a binary blob, or choose when to parse the data.
      #   ##
      #   ## Server pre parses multipart form data by default.
      #   DisablePreParseMultipartForm: !!bool true

      #   ## Logs all errors, including the most frequent
      #   ## 'connection reset by peer', 'broken pipe' and 'connection timeout'
      #   ## errors. Such errors are common in production serving real-world
      #   ## clients.
      #   ##
      #   ## By default the most frequent errors such as
      #   ## 'connection reset by peer', 'broken pipe' and 'connection timeout'
      #   ## are suppressed in order to limit output log traffic.
      #   LogAllErrors: !!bool false

      #   ## Will not log potentially sensitive content in error logs
      #   ##
      #   ## This option is useful for servers that handle sensitive data
      #   ## in the request/response.
      #   ##
      #   ## Server logs all full errors by default.
      #   SecureErrorLogMessage: !!bool false

      #   ## Header names are passed as-is without normalization
      #   ## if this option is set.
      #   ##
      #   ## Disabled header names' normalization may be useful only for proxying
      #   ## incoming requests to other servers expecting case-sensitive
      #   ## header names. See https:##github.com/valyala/fasthttp/issues/57
      #   ## for details.
      #   ##
      #   ## By default request and response header names are normalized, i.e.
      #   ## The first letter and the first letters following dashes
      #   ## are uppercased, while all the other letters are lowercased.
      #   ## Examples:
      #   ##
      #   ##     * HOST -> Host
      #   ##     * content-type -> Content-Type
      #   ##     * cONTENT-lenGTH -> Content-Length
      #   DisableHeaderNamesNormalizing: !!bool false

      #   ## SleepWhenConcurrencyLimitsExceeded is a duration to be slept of if
      #   ## the concurrency limit in exceeded (default [when is 0]: don't sleep
      #   ## and accept new connections immediately).
      #   SleepWhenConcurrencyLimitsExceeded: !!str 0s

      #   ## NoDefaultServerHeader, when set to true, causes the default Server header
      #   ## to be excluded from the Response.
      #   ##
      #   ## The default Server header value is the value of the Name field or an
      #   ## internal default value in its absence. With this option set to true,
      #   ## the only time a Server header will be sent is if a non-zero length
      #   ## value is explicitly provided during a request.
      #   NoDefaultServerHeader: !!bool false

      #   ## NoDefaultDate, when set to true, causes the default Date
      #   ## header to be excluded from the Response.
      #   ##
      #   ## The default Date header value is the current date value. When
      #   ## set to true, the Date will not be present.
      #   NoDefaultDate: !!bool false

      #   ## NoDefaultContentType, when set to true, causes the default Content-Type
      #   ## header to be excluded from the Response.
      #   ##
      #   ## The default Content-Type header value is the internal default value. When
      #   ## set to true, the Content-Type will not be present.
      #   NoDefaultContentType: !!bool false

      #   ## KeepHijackedConns is an opt-in disable of connection
      #   ## close by fasthttp after connections' HijackHandler returns.
      #   ## This allows to save goroutines, e.g. when fasthttp used to upgrade
      #   ## http connections to WS and connection goes to another handler,
      #   ## which will close it when needed.
      #   KeepHijackedConns: !!bool false

      #   ## CloseOnShutdown when true adds a `Connection: close` header when the server is shutting down.
      #   CloseOnShutdown: !!bool true

      #   ## StreamRequestBody enables request body streaming,
      #   ## and calls the handler sooner when given body is
      #   ## larger than the current limit.
      #   StreamRequestBody: !!bool false
```
