package apiv1

import (
	"context"

	cpb "protos-repo/common/common"
	"protos-repo/example/api/v1/user"
	"svc-example/services"

	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

type UserAPI struct {
	svcCtx *services.ServiceContext

	user.UnimplementedUserServer
}

func NewUserAPI(svcCtx *services.ServiceContext) *UserAPI {
	return &UserAPI{svcCtx: svcCtx}
}

func (api *UserAPI) GrpcServiceDesc() *grpc.ServiceDesc { return &user.User_ServiceDesc }
func (api *UserAPI) RestServiceDesc() *rest.ServiceDesc { return &user.UserRestServiceDesc }

// Create persists a new user record.
// In Asjard, this method typically uses 'SetData' to pre-allocate IDs
// and initialize the cache state to prevent early cache-miss storms.
func (api *UserAPI) Create(ctx context.Context, in *user.UserReq) (*cpb.Empty, error) {
	return &cpb.Empty{}, api.svcCtx.Svcs.UserSvc.Create(ctx, in)
}

// Get retrieves a single user by their primary identifier (username/key).
// Workflow: LocalCache -> Redis -> Database.
// Failure to find a record results in a gRPC NOT_FOUND (HTTP 404).
func (api *UserAPI) Get(ctx context.Context, in *cpb.ReqWithName) (*user.UserInfo, error) {
	return api.svcCtx.Svcs.UserSvc.Get(ctx, in)
}

// Update modifies an existing user profile.
// Triggers the 'Delayed Double Delete' strategy via Asjard's Time Wheel
// to ensure consistency across distributed LocalCache nodes.
func (api *UserAPI) Update(ctx context.Context, in *user.UserReq) (*cpb.Empty, error) {
	return &cpb.Empty{}, api.svcCtx.Svcs.UserSvc.Update(ctx, in)
}

// Del removes the user record from the persistent store and
// synchronously purges associated entries from all cache levels.
func (api *UserAPI) Del(ctx context.Context, in *cpb.ReqWithName) (*cpb.Empty, error) {
	return &cpb.Empty{}, api.svcCtx.Svcs.UserSvc.Del(ctx, in)
}

// Search filters users with pagination support.
// Note: Search results are typically cached at the 'Group' level
// with shorter TTLs compared to individual 'Get' records.
func (api *UserAPI) Search(ctx context.Context, in *user.UserSearchReq) (*user.UserList, error) {
	return api.svcCtx.Svcs.UserSvc.Search(ctx, in)
}
