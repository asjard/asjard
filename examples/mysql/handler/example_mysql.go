package handler

import (
	"context"

	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/examples/mysql/model"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc/codes"
)

type MysqlExampleAPI struct {
	table *model.ExampleTable
	pb.UnimplementedHelloServer
}

func NewMysqlExampleAPI() *MysqlExampleAPI {
	return &MysqlExampleAPI{
		table: &model.ExampleTable{},
	}
}

func (MysqlExampleAPI) Bootstrap() error {
	// 数据库初始化
	if err := model.Init(); err != nil {
		return err
	}
	return nil
}

func (MysqlExampleAPI) Shutdown() {
}

func (m MysqlExampleAPI) MysqlExample(ctx context.Context, in *pb.MysqlExampleReq) (*pb.MysqlExampleResp, error) {
	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is must")
	}
	return m.table.AddOrUpdate(ctx, in)
}

func (MysqlExampleAPI) RestServiceDesc() *rest.ServiceDesc {
	return &pb.HelloRestServiceDesc
}
