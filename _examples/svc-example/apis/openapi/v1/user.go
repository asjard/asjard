package openapiv1

import (
	"context"

	cpb "protos-repo/common/common"
	"protos-repo/example/openapi/v1/user"
	"svc-example/datas"
	"svc-example/services"

	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

type UserAPI struct {
	*services.ServiceContext

	user.UnimplementedUserServer
}

func NewUserAPI(svcCtx *services.ServiceContext) *UserAPI {
	return &UserAPI{ServiceContext: svcCtx}
}

func (api *UserAPI) GrpcServiceDesc() *grpc.ServiceDesc { return &user.User_ServiceDesc }
func (api *UserAPI) RestServiceDesc() *rest.ServiceDesc { return &user.UserRestServiceDesc }

// Get retrieves a single user by their primary identifier (username/key).
// Workflow: LocalCache -> Redis -> Database.
// Failure to find a record results in a gRPC NOT_FOUND (HTTP 404).
func (api UserAPI) Get(ctx context.Context, in *cpb.ReqWithName) (*user.UserInfo, error) {
	record, err := api.UserSvc.Get(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	return api.userInfo(record), nil
}

func (api *UserAPI) userInfo(t *datas.User) *user.UserInfo {
	return &user.UserInfo{
		Username: t.Username,
		Age:      t.Age,
	}
}
