详细示例参考[这里](https://github.com/asjard/examples/blob/develop/_examples/svc-example/services/user.go)

## 使用

```go
package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/cache"
	"github.com/asjard/asjard/pkg/protobuf/requestpb"
	"github.com/asjard/asjard/pkg/stores"
	"github.com/asjard/asjard/pkg/stores/xgorm"
	pb "github.com/asjard/examples/protobuf/mysqlpb"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type ExampleTable struct {
	Id        int64  `gorm:"column:id;type:INT(20);primaryKey;autoIncrement"`
	Name      string `gorm:"column:name;type:VARCHAR(20);uniqueIndex"`
	Age       uint32 `gorm:"column:age;type:INT"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExampleModel struct {
	stores.Model
	*ExampleTable
	kvCache *cache.CacheRedis
}

// TableName 数据库表名
func (ExampleTable) TableName() string {
	return "example_table"
}

// ModelName 全局唯一的表名
func (ExampleTable) ModelName() string {
	return "example_database_example_table"
}

// NewExampleModel 模型初始化
// 用XXXModel去复写XXXTable的方法
func NewExampleModel() *ExampleModel {
	exampleModel := &ExampleModel{
		ExampleTable: &ExampleTable{},
	}
	bootstrap.AddBootstrap(exampleModel)
	return exampleModel
}

// Bootstrap 缓存初始化
func (model *ExampleModel) Bootstrap() (err error) {
	localCache, err := cache.NewLocalCache(model.ExampleTable)
	if err != nil {
		return err
	}
	model.kvCache, err = cache.NewRedisKeyValueCache(model.ExampleTable,
		cache.WithLocalCache(localCache))
	if err != nil {
		return err
	}
	return nil
}

// Shutdown .
func (model *ExampleModel) Shutdown() {}

func (model *ExampleModel) Create(ctx context.Context, in *pb.CreateOrUpdateReq) error {
	// 需要删除删除缓存
	// 创建需要提前分配缓存Key
	// 如果是按照主键ID缓存的，提前生成主键ID，创建记录使用提前生成的主键ID创建记录
	if err := model.SetData(ctx,
		model.kvCache.WithGroup(model.searchGroup()).WithKey(model.getCacheKey(in.Name)),
		func() error {
			return model.ExampleTable.Create(ctx, in)
		}); err != nil {
		return err
	}
	return nil
}

func (model *ExampleModel) Update(ctx context.Context, in *pb.CreateOrUpdateReq) (*pb.ExampleInfo, error) {
	if err := model.SetData(ctx,
		model.kvCache.WithGroup(model.searchGroup()).WithKey(model.getCacheKey(in.Name)),
		func() error {
			if _, err := model.ExampleTable.Update(ctx, in); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return model.Get(ctx, &pb.ReqWithName{Name: in.Name})
}

func (model *ExampleModel) Get(ctx context.Context, in *pb.ReqWithName) (*pb.ExampleInfo, error) {
	var resp pb.ExampleInfo
	if err := model.GetData(ctx,
		&resp,
		model.kvCache.WithKey(model.getCacheKey(in.Name)),
		func() (any, error) {
			return model.ExampleTable.Get(ctx, in)
		}); err != nil || resp.Id == 0 {
		return nil, status.Errorf(codes.NotFound, "record not found")
	}
	return &resp, nil
}

func (model *ExampleModel) Search(ctx context.Context, in *pb.SearchReq) (*pb.ExampleList, error) {
	var result pb.ExampleList
	if err := model.GetData(ctx, &result,
		model.kvCache.WithKey(model.searchCacheKey(in)).WithGroup(model.searchGroup()),
		func() (any, error) {
			return model.ExampleTable.Search(ctx, in)
		}); err != nil {
		return nil, err
	}
	return &result, nil
}

func (model *ExampleModel) Del(ctx context.Context, in *pb.ReqWithName) error {
	return model.SetData(ctx,
		model.kvCache.WithKey(model.getCacheKey(in.Name)).WithGroup(model.searchGroup()),
		func() error {
			return model.ExampleTable.Del(ctx, in)
		})
}

func (model *ExampleModel) searchGroup() string {
	return "search"
}

func (model *ExampleModel) searchCacheKey(in *pb.SearchReq) string {
	return fmt.Sprintf("search:%d:%d:%s", in.Page, in.Size, in.Sort)
}

func (model *ExampleModel) getCacheKey(name string) string {
	return fmt.Sprintf("info:%s", name)
}

func (t ExampleTable) Create(ctx context.Context, in *pb.CreateOrUpdateReq) error {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Where("name=?", in.Name).First(&ExampleTable{}).Error; err == nil {
		return status.Errorf(codes.AlreadyExists, "record %s already exist", in.Name)
	}

	if err := db.Create(&ExampleTable{
		Name: in.Name,
		Age:  in.Age,
	}).Error; err != nil {
		logger.Error("create fail", "name", in.Name, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t ExampleTable) Update(ctx context.Context, in *pb.CreateOrUpdateReq) (*pb.ExampleInfo, error) {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := t.Get(ctx, &pb.ReqWithName{Name: in.Name}); err != nil {
		return nil, err
	}
	if err := db.Model(&ExampleTable{}).Where("name=?", in.Name).Updates(&ExampleTable{
		Name: in.Name,
		Age:  in.Age,
	}).Error; err != nil {
		logger.Error("update record fail", "name", in.Name, "err", err)
		return nil, status.InternalServerError()
	}
	return t.Get(ctx, &pb.ReqWithName{Name: in.Name})
}

func (t ExampleTable) Get(ctx context.Context, in *pb.ReqWithName) (*pb.ExampleInfo, error) {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return nil, err
	}
	var record ExampleTable
	if err := db.Where("name=?", in.Name).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "record not found")
		}
		logger.Error("get record fail", "name", in.Name, "err", err)
		return nil, status.InternalServerError()
	}
	return record.info(), nil
}

func (t ExampleTable) Search(ctx context.Context, in *pb.SearchReq) (*pb.ExampleList, error) {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return nil, err
	}
	var total int64
	db.Model(&ExampleTable{}).Count(&total)
	records := make([]ExampleTable, 0, in.Size)
	db.Scopes(requestpb.ReqWithPageGormScope(in.Page, in.Size, in.Sort)).Find(&records)
	exampleInfos := make([]*pb.ExampleInfo, 0, in.Size)
	for _, record := range records {
		exampleInfos = append(exampleInfos, record.info())
	}
	return &pb.ExampleList{
		Total: int32(total),
		List:  exampleInfos,
	}, nil
}

func (t ExampleTable) Del(ctx context.Context, in *pb.ReqWithName) error {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Where("name=?", in.Name).Delete(&ExampleTable{}).Error; err != nil {
		logger.Error("delete record fail", "name", in.Name, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t ExampleTable) info() *pb.ExampleInfo {
	return &pb.ExampleInfo{
		Id:        t.Id,
		Name:      t.Name,
		Age:       t.Age,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
}

```
