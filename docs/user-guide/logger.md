## 配置

```yaml
asjard:
  ## log configuration
  logger:
    ## supports: DEBUG,INFO,WARN,ERROR,Case insensitive
    # level: INFO
    ## log format, supports: text,json
    # format: json
    ## log file path
    # filePath: /dev/stdout
    ## Log explosion prevention configuration
    ## File size (in MB)
    # maxSize: 100
    ## Maximum number of days to keep a file.
    # maxAge: 0
    ## Maximum number of backups.
    # maxBackups: 10
    ## Whether to compress
    # compress: true
    ## Log desensitization (global slog.Handler wrapper, on by default)
    desensitize:
      ## disable field / content masking
      # disabled: false
      ## replacement text for masked values
      # replacement: "***"
      ## global keep: keepPrefix + replacement + keepSuffix; if length <= prefix+suffix, use replacement only
      # keepPrefix: 0
      # keepSuffix: 0
      ## recursion depth for nested values
      # maxDepth: 8
      ## sensitive field names and/or content patterns
      ## per-key keepPrefix/keepSuffix/replacement: 0/empty falls back to global when global keep > 0
      # keys:
      # - name: password
      # - name: mobile
      #   keepPrefix: 3
      #   keepSuffix: 4
      # - name: phoneInText
      #   pattern: '1[3-9]\d{9}'
      #   keepPrefix: 3
      #   keepSuffix: 4
    ## access log configuration
    accessLog:
      ## enable access log
      # enabled: false
      ## format: [protocol://]{fullMethod}
      ## skip specify protocol and method: grpc:///api.v1.hello.Hello/Call
      ## skip a protocol-independent method: /api.v1.hello.Hello/Call
      ## skip all grpc proto: grpc
      # skipMethods:
      # - grpc
      ## other configurations same as asjard.logger
    ## gorm log configuration
    gorm:
      # ignoreRecordNotFoundError: false
      # slowThreshold: 200ms
      ## other configurations same as asjard.logger
    ## banner log
    banner:
      ## if enabled it will be print banner message after server start.
      # enabled: true
```

## 使用

```go
import "github.com/asjard/asjard/v2/core/logger"

func (api *SampleAPI) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	logger.Debug("say hello request", "name", name)
	// 或者带链路追踪的logger
	logger.L(ctx).Debug("say hello request", "name", name)
	return &pb.HelloReply{
		Message: "hello " + in.Name,
	}, nil
}
```

## 日志脱敏

在 `slog.Handler` 层对所有日志属性递归脱敏（含 accessLog、gorm 等通过 `NewSlogHandler` 创建的 handler）。**默认开启**，并随 `asjard.logger` / `asjard.logger.desensitize` 配置热更新。

- **replacement**：脱敏占位符，默认 `***`
- **keepPrefix / keepSuffix**：全局前/后保留位数（按 rune）；格式为 `前 + replacement + 后`；默认为前3后4，或字符串长度 `<= keepPrefix+keepSuffix` 时，整段替换为 `replacement`
- **maxDepth**：嵌套结构递归深度，默认 `8`
- **keys**：
  - `name`：按字段名 / Header 名匹配（大小写不敏感）
  - `pattern`：对字符串叶子做正则子串替换（配置了 `pattern` 时按内容规则处理，不再按 `name` 注册字段规则）
  - 每条可单独配置 `keepPrefix` / `keepSuffix` / `replacement`；未配置或 keep 值为 `0` 且全局 keep `> 0` 时，回落使用全局配置, 小于0时则为0
  - 用户 `keys` 优先于内置 `builtInKeys`（同名字段用户配置覆盖内置）

内置敏感字段名包括：`password`、`passwd`、`token`、`authorization`、`cookie`、`mobile`、`phone`、`id_card`、`bank_card` 等。

关闭示例：

```yaml
asjard:
  logger:
    desensitize:
      disabled: true
```

自定义示例：

```yaml
asjard:
  logger:
    desensitize:
      replacement: "***"
      keepPrefix: 1
      keepSuffix: 1
      keys:
        - name: mobile
          keepPrefix: 3
          keepSuffix: 4
        - name: phoneInText
          pattern: '1[3-9]\d{9}'
          keepPrefix: 3
          keepSuffix: 4
```
