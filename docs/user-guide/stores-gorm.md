## 配置

```yaml
asjard:
  ## 数据相关配置
  stores:
    ## gorm数据库相关配置
    gorm:
      ## 数据库列表
      dbs:
        ## default数据库配置
        default:
          ## 此字段受cipherName保护
          dsn: root:my-secret-pw@tcp(127.0.0.1:3306)/exmple-database?charset=utf8&parseTime=True&loc=Local
          ## 数据库驱动
          ## mysql, postgres,sqlite,sqlserver,clickhouse
          ## ref: https://gorm.io/zh_CN/docs/connecting_to_the_database.html#PostgreSQL
          driver: mysql
          ## 加密组件名称
          ## 不为空则使用加密组件解密后连接数据库
          cipherName: ""
          ## 加密组件参数, map[string]any类型
          cipherParams: {}
          ## 驱动自定义配置
          options:
            ## 继承asjard.stores.gorm.options
            ## 自定义驱动名称
            ## ref: https://gorm.io/docs/connecting_to_the_database.html#Customize-Driver
            # driverName: ""
      ## 数据库连接配置
      options:
        # maxIdleConns: 10
        # maxOpenConns: 1001
        # connMaxIdleTime: 10
        # connMaxLifeTime: 2h
        # debug: false
        # skipInitializeWithVersion: false
        # skipDefaultTransaction: false
        # traceable: false
        # metricsable: false
        # translateError: false
```

## 使用

```go
import "github.com/asjard/asjard/pkg/stores/xgorm"

// 使用默认客户端
db, err := xgorm.DB(context.Background())
if err != nil {
	return err
}

// 自定义客户端
// 前提是你需要配置asjrd.stores.gorm.dbs.xxx
db, err := xgorm.DB(context.Background(), xgorm.WithConnName("xxx"))
```
