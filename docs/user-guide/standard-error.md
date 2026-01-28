## 约定

- 约定所有**对外**暴露的错误使用`github.com/asjard/asjard/core/status`提供的`Error`和`Errorf`方法

### 错误码约定

格式: XXX_YYY_ZZZ

其中:

- XXX: 代表系统码, 固定三位数字, 例如: 100, 101, 102等， 可通过`asjard.service.instance.systemCode`配置
- YYY: 代表[HTTP状态码](https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Status)， 固定三位数字, 例如: 400, 404, 500等
- ZZZ: 代表错误码, 不定长数字, 例如: 201,202,20001等, 其中:
  - `<=1 ZZZ <= 17`为全局共享错误码,任何系统都可使用
  - `18 <= ZZZ <= 200`为框架保留错误码，业务系统请勿使用

## 使用示例

```go

import （
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/database/mysql"
）

const (

	// 自定义XXX Not Found错误码
	CustomeXXXNotFoundErrorCode = 404_201
)

var (
	// 自定义错误，注意这里需要使用匿名函数的方式定义
	// 里面包含获取systemCode逻辑，如果定义为全局错误则会出现没有加载配置文件问题
	CustomeXXXError = func() error {return status.Error(CustomeXXXNotFoundErrorCode, "define error as a variable")}
)

func(api XXXAPI) YYY(ctx context, in *pb.Req) (*pb.Resp, error) {
	if in.Name == "" {
		// 此处返回的是全局共享错误码
		return nil, status.Error(codes.InvalidArgument, "name is must")
	}

	db, err := mysql.DB(ctx)
	if err != nil {
		// 此处返回的是框架保留的错误码
		// 同 status.DatabaseNotFoundError
		return err
	}

	var record ExampleTable
	if err := db.Where("name=?", in.Name).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 此处返回的是自定义错误
			return nil, status.Errorf(CustomeXXXNotFoundErrorCode, "recode %s not found", in.Name)
		}
	}
	return &pb.Resp{}, nil
}
```
