## accessLog

支持协议: all

```yaml
asjard:
  logger:
    accessLog:
      ## 是否开启accessLog
      enabled: true
      ## 配置格式: [protocol://]{fullMethod}
      ## 例如grpc协议的某个方法: grpc:///api.v1.hello.Hello/Call
      ## 或者协议无关的某个方法: /api.v1.hello.Hello/Call
      ## 拦截协议的所有方法: grpc
      skipMethods:
        - grpc
```

## i18n

支持协议: rest

国际化配置存放路径, 优先读取`ASJARD_I18N_DIR`环境变量
如果为空读取`ASJARD_HOME_DIR`
如果为空则为`可执行程序`所在路径的`locals`目录

文件格式为`{lang}.json`， 其中`lang`为语言

文件内容格式为:

```json
{
  "123": {
    "prompt": "错误提示"
    "doc": "可以自行处理这个错误的文档地址"
  }
}
```

其中`123`为错误码, 只包含`系统码`和`错误码`

## ratelimiter

```yaml
asjard:
  ## 拦截器相关配置
  interceptors:
    ## 服务端拦截器相关配置
    server:
      ## 限速器配置
      rateLimiter:
        ## 是否开启限速
        # enabled: false
        ## 每秒最多多少个请求
        ## <0表示不限制
        # limit: -1
        ## 桶容量大小,小于0则为limit值
        # burst: -1
        ## 单独方法的限速配置
        methods:
          ## 方法名称
          ## [{protocol}://]{method}
          ## 所有协议健康检查限速每秒10个请求
          # - name: /asjard.api.health.Health/Check
          #   limit: 10
          #   burst: 10
          ## grpc协议的/api.v1.server.Server/Hello限速每秒20个请求
          # - name: grpc:///api.v1.server.Server/Hello
          #   limit: 20
          #   burst: 20
```
