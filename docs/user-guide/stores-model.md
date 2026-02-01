详细示例参考[这里](https://github.com/asjard/asjard/blob/develop/_examples/svc-example/services/user.go)

## 使用

```go
package services

import (
	"context"
	"fmt"
	"sync"

	cpb "protos-repo/common/common"
	"protos-repo/common/xcodes"
	"protos-repo/example/api/v1/user"
	"svc-example/datas"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/cache"
	"github.com/asjard/asjard/pkg/stores"
	"google.golang.org/grpc/codes"
)

type UserSvc struct {
	datas.User
	stores.Model

	kvCache *cache.CacheRedis
}

var (
	userSvc     *UserSvc
	userSvcOnce sync.Once
)

func NewUserSvc() *UserSvc {
	userSvcOnce.Do(func() {
		userSvc = &UserSvc{}
		bootstrap.AddBootstrap(userSvc)
	})
	return userSvc
}

func (s *UserSvc) Start() error {
	localCache, err := cache.NewLocalCache(s)
	if err != nil {
		return err
	}
	s.kvCache, err = cache.NewRedisKeyValueCache(s, cache.WithLocalCache(localCache))
	if err != nil {
		return err
	}

	return nil
}
func (s *UserSvc) Stop() {}

func (s *UserSvc) Create(ctx context.Context, in *user.UserReq) error {
	record, err := s.Get(ctx, &cpb.ReqWithName{Name: in.Username})
	if err == nil && record.UserId != 0 {
		return status.Errorf(codes.Code(xcodes.ERR_USER_EUSE_EXIST), "user '%s' already exist", in.Username)
	}
	return s.SetData(ctx, func() error {
		return s.User.Create(ctx, in)
	}, s.kvCache.WithKey(s.usernameCacheKey(in.Username)).WithGroup(s.searchCacheGroupKey()))
}

func (s *UserSvc) Update(ctx context.Context, in *user.UserReq) error {
	if _, err := s.Get(ctx, &cpb.ReqWithName{Name: in.Username}); err != nil {
		return err
	}
	return s.SetData(ctx, func() error {
		return s.User.Update(ctx, in)
	}, s.kvCache.WithKey(s.usernameCacheKey(in.Username)).WithGroup(s.searchCacheGroupKey()))
}

func (s *UserSvc) UpdateCardNum(ctx context.Context, username string, num int) error {
	if _, err := s.Get(ctx, &cpb.ReqWithName{Name: username}); err != nil {
		return err
	}
	return s.SetData(ctx, func() error {
		return s.User.UpdateCardNum(ctx, username, num)
	}, s.kvCache.WithKey(s.usernameCacheKey(username)).WithGroup(s.searchCacheGroupKey()))
}

func (s *UserSvc) Get(ctx context.Context, in *cpb.ReqWithName) (*user.UserInfo, error) {
	var record user.UserInfo
	if err := s.GetData(ctx,
		&record,
		s.kvCache.WithKey(s.usernameCacheKey(in.Name)), func() (any, error) {
			return s.User.Get(ctx, in)
		}); err != nil {
		return nil, err
	}
	if record.UserId == 0 {
		return nil, status.Errorf(codes.Code(xcodes.ERR_USER_EUSE_NOT_FOUND), "user '%s' not found", in.Name)
	}
	return &record, nil
}

func (s *UserSvc) Del(ctx context.Context, in *cpb.ReqWithName) error {
	if _, err := s.Get(ctx, in); err != nil {
		return err
	}
	return s.SetData(ctx, func() error {
		return s.User.Del(ctx, in)
	}, s.kvCache.WithKey(s.usernameCacheKey(in.Name)).WithGroup(s.searchCacheGroupKey()))
}

func (s *UserSvc) Search(ctx context.Context, in *user.UserSearchReq) (*user.UserList, error) {
	var record user.UserList
	if err := s.GetData(ctx,
		&record,
		s.kvCache.WithKey(s.searchCacheKey(in)).WithGroup(s.searchCacheGroupKey()),
		func() (any, error) {
			return s.User.Search(ctx, in)
		}); err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *UserSvc) usernameCacheKey(username string) string {
	return fmt.Sprintf("username:%s", username)
}
func (s *UserSvc) searchCacheGroupKey() string {
	return "search"
}

func (s *UserSvc) searchCacheKey(in *user.UserSearchReq) string {
	return fmt.Sprintf("search:%d:%d:%s:%s", in.Page, in.Size, in.Sort, in.Keywords)
}

```
