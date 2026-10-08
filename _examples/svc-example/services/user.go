package services

import (
	"context"
	"fmt"

	"protos-repo/common/xcodes"
	"svc-example/datas"
	"svc-example/datas/models"

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

func NewUserSvc() *UserSvc {
	s := &UserSvc{}
	bootstrap.AddBootstrap(s)
	return s
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

func (s *UserSvc) ModelName() string { return s.TableName() }

func (s *UserSvc) Create(ctx context.Context, in *datas.User) error {
	record, err := s.Get(ctx, in.Username)
	if err == nil && record.ID != 0 {
		return status.Errorf(codes.Code(xcodes.ERR_USER_EUSE_EXIST), "user '%s' already exist", in.Username)
	}
	return s.SetData(ctx, func() error {
		return s.User.Create(ctx, in)
	}, s.kvCache.WithKey(s.usernameCacheKey(in.Username)).WithGroup(s.searchCacheGroupKey()))
}

func (s *UserSvc) Update(ctx context.Context, in *datas.User) error {
	if _, err := s.Get(ctx, in.Username); err != nil {
		return err
	}
	return s.SetData(ctx, func() error {
		return s.User.Update(ctx, in)
	}, s.kvCache.WithKey(s.usernameCacheKey(in.Username)).WithGroup(s.searchCacheGroupKey()))
}

func (s *UserSvc) UpdateCardNum(ctx context.Context, username string, num int) error {
	if _, err := s.Get(ctx, username); err != nil {
		return err
	}
	return s.SetData(ctx, func() error {
		return s.User.UpdateCardNum(ctx, username, num)
	}, s.kvCache.WithKey(s.usernameCacheKey(username)).WithGroup(s.searchCacheGroupKey()))
}

func (s *UserSvc) Get(ctx context.Context, username string) (*datas.User, error) {
	var record datas.User
	if err := s.GetData(ctx,
		&record,
		s.kvCache.WithKey(s.usernameCacheKey(username)), func() (any, error) {
			return s.User.Get(ctx, username)
		}); err != nil {
		return nil, err
	}
	if record.ID == 0 {
		return nil, status.Errorf(codes.Code(xcodes.ERR_USER_EUSE_NOT_FOUND), "user '%s' not found", username)
	}
	return &record, nil
}

func (s *UserSvc) Del(ctx context.Context, username string) error {
	if _, err := s.Get(ctx, username); err != nil {
		return err
	}
	return s.SetData(ctx, func() error {
		return s.User.Del(ctx, username)
	}, s.kvCache.WithKey(s.usernameCacheKey(username)).WithGroup(s.searchCacheGroupKey()))
}

type userSearchResult struct {
	Total   int32
	Records []*datas.User
}

func (s *UserSvc) Search(ctx context.Context, in *models.UserSearchReq) (int32, []*datas.User, error) {
	var result userSearchResult
	if err := s.GetData(ctx,
		&result,
		s.kvCache.WithKey(s.searchCacheKey(in)).WithGroup(s.searchCacheGroupKey()),
		func() (any, error) {
			total, records, err := s.User.Search(ctx, in)
			if err != nil {
				return nil, err
			}
			return &userSearchResult{Total: total, Records: records}, nil
		}); err != nil {
		return 0, nil, err
	}
	return result.Total, result.Records, nil
}

func (s *UserSvc) usernameCacheKey(username string) string {
	return fmt.Sprintf("username:%s", username)
}
func (s *UserSvc) searchCacheGroupKey() string {
	return "search"
}

func (s *UserSvc) searchCacheKey(in *models.UserSearchReq) string {
	return fmt.Sprintf("search:%d:%d:%s:%s", in.Page, in.Size, in.Sort, in.Keywords)
}
