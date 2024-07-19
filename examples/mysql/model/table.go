package model

import (
	"context"
	"errors"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	"github.com/asjard/asjard/pkg/database/mysql"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type ExampleTable struct {
	ID      int64          `gorm:"column:id;type:INT(20);primaryKey;autoIncrement"`
	Name    string         `gorm:"column:name;type:VARCHAR(20);unique_index"`
	Age     uint32         `gorm:"column:age;type:INT"`
	Volumes pq.StringArray `gorm:"column:volumes;type:MEDIUMTEXT;comment:存储"`
	Envs    pq.StringArray `gorm:"column:envs;type:MEDIUMTEXT;comment:环境变量"`
	// Ports     pq.Int64Array  `gorm:"column:ports:type:MEDIUMTEXT;comment:端口列表"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ExampleTable) TableName() string {
	return "example_table"
}

// 添加数据
func (t ExampleTable) AddOrUpdate(ctx context.Context, in *pb.MysqlExampleReq) (*pb.MysqlExampleResp, error) {
	db, err := mysql.DB(ctx)
	if err != nil {
		return nil, err
	}
	// 不存在创建，存在更新
	if err = db.Where("name=?", in.Name).First(&ExampleTable{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&ExampleTable{
				Name: in.Name,
				Age:  in.Age,
			}).Error; err != nil {
				logger.Error("create record fail", "req", in, "err", err.Error())
				return nil, status.Error(codes.Internal, "Internal server error")
			}
			return t.getByName(ctx, db, in.Name)
		}
	}
	if err := db.Where("name=?", in.Name).Updates(&ExampleTable{
		Age: in.Age,
	}).Error; err != nil {
		logger.Error("update record fail", "name", in.Name, "err", err.Error())
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return t.getByName(ctx, db, in.Name)
}

func (ExampleTable) getByName(_ context.Context, db *gorm.DB, name string) (*pb.MysqlExampleResp, error) {
	var table ExampleTable
	if err := db.Where("name=?", name).First(&table).Error; err != nil {
		logger.Error("get record not found", "name", name, "err", err.Error())
		return nil, status.Error(codes.NotFound, "record not found")
	}
	return table.info(), nil
}

func (t ExampleTable) info() *pb.MysqlExampleResp {
	return &pb.MysqlExampleResp{
		Id:        int64(t.ID),
		Name:      t.Name,
		Age:       t.Age,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
}
