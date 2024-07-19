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

func (m MysqlExampleAPI) MysqlExample(ctx context.Context, in *pb.MysqlExampleReq) (*pb.MysqlExampleResp, error) {
	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is must")
	}
	return m.table.AddOrUpdate(ctx, in)
}

func (MysqlExampleAPI) RestServiceDesc() *rest.ServiceDesc {
	return &pb.HelloRestServiceDesc
}
