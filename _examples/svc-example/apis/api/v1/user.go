package apiv1

import (
	"context"

	cpb "protos-repo/common/common"
	"protos-repo/example/api/v1/user"
	"svc-example/datas"
	"svc-example/datas/models"
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

// Create persists a new user record.
// In Asjard, this method typically uses 'SetData' to pre-allocate IDs
// and initialize the cache state to prevent early cache-miss storms.
func (api *UserAPI) Create(ctx context.Context, in *user.UserReq) (*cpb.Empty, error) {
	return &cpb.Empty{}, api.UserSvc.Create(ctx, &datas.User{
		Username: in.Username,
		Age:      in.Age,
	})
}

// Get retrieves a single user by their primary identifier (username/key).
// Workflow: LocalCache -> Redis -> Database.
// Failure to find a record results in a gRPC NOT_FOUND (HTTP 404).
func (api *UserAPI) Get(ctx context.Context, in *cpb.ReqWithName) (*user.UserInfo, error) {
	record, err := api.UserSvc.Get(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	return api.userInfo(record), nil
}

// Update modifies an existing user profile.
// Triggers the 'Delayed Double Delete' strategy via Asjard's Time Wheel
// to ensure consistency across distributed LocalCache nodes.
func (api *UserAPI) Update(ctx context.Context, in *user.UserReq) (*cpb.Empty, error) {
	return &cpb.Empty{}, api.UserSvc.Update(ctx, &datas.User{
		Username: in.Username,
		Age:      in.Age,
	})
}

// Del removes the user record from the persistent store and
// synchronously purges associated entries from all cache levels.
func (api *UserAPI) Del(ctx context.Context, in *cpb.ReqWithName) (*cpb.Empty, error) {
	return &cpb.Empty{}, api.UserSvc.DoWithTransaction(ctx, func(ctx context.Context) error {
		if err := api.UserCreditCardSvc.RemoveByUser(ctx, in.Name); err != nil {
			return err
		}
		return api.UserSvc.Del(ctx, in.Name)
	})
}

// Search filters users with pagination support.
// Note: Search results are typically cached at the 'Group' level
// with shorter TTLs compared to individual 'Get' records.
func (api *UserAPI) Search(ctx context.Context, in *user.UserSearchReq) (*user.UserList, error) {
	total, records, err := api.UserSvc.Search(ctx, &models.UserSearchReq{
		Page:     in.Page,
		Size:     in.Size,
		Sort:     in.Sort,
		Keywords: in.Keywords,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*user.UserInfo, len(records))
	for idx, item := range records {
		list[idx] = api.userInfo(item)
	}
	return &user.UserList{
		Total: total,
		List:  list,
	}, nil
}

// User add a credit card
func (api *UserAPI) AddCreditCard(ctx context.Context, in *user.UserCreditCardReq) (*cpb.Empty, error) {
	return &cpb.Empty{}, api.UserSvc.DoWithTransaction(ctx, func(ctx context.Context) error {
		if err := api.UserCreditCardSvc.Add(ctx, &datas.UserCreditCard{Username: in.Username, Number: in.Number}); err != nil {
			return err
		}
		return api.UserSvc.UpdateCardNum(ctx, in.Username, 1)
	})
}

// User remove a credit card
func (api *UserAPI) RemoveCreditCard(ctx context.Context, in *user.UserCreditCardReq) (*cpb.Empty, error) {
	return &cpb.Empty{}, api.UserCreditCardSvc.DoWithTransaction(ctx, func(ctx context.Context) error {
		if err := api.UserCreditCardSvc.Remove(ctx, &datas.UserCreditCard{Username: in.Username, Number: in.Number}); err != nil {
			return err
		}
		return api.UserSvc.UpdateCardNum(ctx, in.Username, -1)
	})
}

// get user credit card info
func (api *UserAPI) GetCreditCard(ctx context.Context, in *user.UserCreditCardReq) (*user.UserCreditCardInfo, error) {
	record, err := api.UserCreditCardSvc.Get(ctx, &datas.UserCreditCard{Username: in.Username, Number: in.Number})
	if err != nil {
		return nil, err
	}
	return api.cardInfo(record), nil
}

// Get all credit cards under user
func (api *UserAPI) SearchCreditCard(ctx context.Context, in *cpb.ReqWithName) (*user.UserCreditCardList, error) {
	total, records, err := api.UserCreditCardSvc.Search(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	list := make([]*user.UserCreditCardInfo, len(records))
	for idx, item := range records {
		list[idx] = api.cardInfo(item)
	}
	return &user.UserCreditCardList{
		Total: total,
		List:  list,
	}, nil
}

func (api *UserAPI) userInfo(t *datas.User) *user.UserInfo {
	return &user.UserInfo{
		UserId:   t.ID,
		Username: t.Username,
		Age:      t.Age,
		CardNums: t.CardNums,
	}
}

func (api *UserAPI) cardInfo(t *datas.UserCreditCard) *user.UserCreditCardInfo {
	return &user.UserCreditCardInfo{
		Number: t.Number,
	}
}
