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
      ## REST uses net/http.Server. See https://pkg.go.dev/net/http#Server
      # options:
      #   ## ReadTimeout is the maximum duration for reading the entire
      #   ## request, including the body. Default: 30s.
      #   readTimeout: !!str 30s

      #   ## ReadHeaderTimeout is the amount of time allowed to read
      #   ## request headers. Default: 10s.
      #   readHeaderTimeout: !!str 10s

      #   ## WriteTimeout is the maximum duration before timing out
      #   ## writes of the response. Default: unlimited (0).
      #   writeTimeout: !!str 0s

      #   ## IdleTimeout is the maximum amount of time to wait for the
      #   ## next request when keep-alive is enabled. Default: 60s.
      #   idleTimeout: !!str 60s

      #   ## MaxHeaderBytes controls the maximum number of bytes the
      #   ## server will read parsing the request header.
      #   maxHeaderBytes: !!int 0

      #   ## Maximum request body size.
      #   ## Default: 20 * 1024 * 1024
      #   maxRequestBodySize: !!int 20971520

      #   ## Whether to disable keep-alive connections.
      #   ## Default: false
      #   disableKeepalive: !!bool false
```
