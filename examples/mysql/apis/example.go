package apis

import (
	"context"

	"github.com/asjard/asjard/examples/mysql/model"
	"github.com/asjard/asjard/examples/mysql/model/xerr"
	pb "github.com/asjard/asjard/examples/protobuf/mysql"
	"github.com/asjard/asjard/pkg/protobuf/requestpb"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/go-playground/validator"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ExampleAPI struct {
	exampleModel *model.ExampleModel
	pb.UnimplementedMysqlServer
}

func NewExampleAPI() *ExampleAPI {
	return &ExampleAPI{
		exampleModel: model.NewExampleModel(),
	}
}

// 创建
func (api *ExampleAPI) Create(ctx context.Context, in *pb.CreateOrUpdateReq) (*emptypb.Empty, error) {
	if err := api.isNameValid(in.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, api.exampleModel.Create(ctx, in)
}

// 更新
func (api *ExampleAPI) Update(ctx context.Context, in *pb.CreateOrUpdateReq) (*pb.ExampleInfo, error) {
	if err := api.isNameValid(in.Name); err != nil {
		return nil, err
	}
	return api.exampleModel.Update(ctx, in)
}

// 获取详情
func (api *ExampleAPI) Get(ctx context.Context, in *pb.ReqWithName) (*pb.ExampleInfo, error) {
	return api.exampleModel.Get(ctx, in)
}

// 查询
func (api *ExampleAPI) Search(ctx context.Context, in *pb.SearchReq) (*pb.ExampleList, error) {
	reqWithPage := &requestpb.ReqWithPage{
		Page: in.Page,
		Size: in.Size,
		Sort: in.Sort,
	}
	if err := reqWithPage.IsValid(20, []string{"created_at", "updated_at"}); err != nil {
		return nil, err
	}
	in.Page = reqWithPage.Page
	in.Size = reqWithPage.Size
	in.Sort = reqWithPage.Sort
	return api.exampleModel.Search(ctx, in)
}

// 删除
func (api *ExampleAPI) Del(ctx context.Context, in *pb.ReqWithName) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, api.exampleModel.Del(ctx, in)
}

func (api *ExampleAPI) RestServiceDesc() *rest.ServiceDesc {
	return &pb.MysqlRestServiceDesc
}

func (api *ExampleAPI) isNameValid(name string) error {
	v := validator.New()
	if err := v.Var(name, "required,max=20"); err != nil {
		return xerr.ErrInvalidName
	}
	return nil
}
