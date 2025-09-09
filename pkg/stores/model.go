package stores

import (
	"context"
	"reflect"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"golang.org/x/sync/singleflight"
)

// Model provider some methods that manager data in database and cache
type Model struct {
	sg singleflight.Group
}

// Modeler is a interface that who want need mangager data in database and cache to implement
type Modeler interface {
	// unique model name, like: database_name_table_name
	ModelName() string
}

const (
	// DefaultSingleflightKey default single flight key
	DefaultSingleflightKey = "default"
)

// GetData from database or cache.
// get data from cache first
// if cache not enabled or not found in cache
// call get function get data from database
func (m *Model) GetData(ctx context.Context, out any, cache Cacher, get func() (any, error)) (err error) {
	if out == nil {
		logger.Error("GetData out is nil")
		return status.InternalServerError()
	}
	if cache == nil || !cache.Enabled() || cache.Key() == "" {
		result, err := get()
		if err != nil {
			return err
		}
		return m.copy(result, out)
	}
	// get data from cache
	fromCurrent, err := cache.Get(ctx, cache.Key(), out)
	if err != nil {
		// get data from database
		result, err, _ := m.sg.Do(cache.Key(), get)
		if err != nil {
			// if data not found from database
			// set a empty data in cache
			if rerr := cache.Set(ctx, cache.Key(), out, cache.EmptyExpiresIn()); rerr != nil {
				logger.Error("set empty into cache fail", "err", rerr)
			}
			return err
		}
		// update cache
		if err := cache.Set(ctx, cache.Key(), result, cache.ExpiresIn()); err != nil {
			logger.Error("set cache fail", "key", cache.Key(), "err", err)
		}
		return m.copy(result, out)
	}
	// refresh cache expire time
	if fromCurrent && cache.AutoRefresh() {
		if err := cache.Refresh(ctx, cache.Key(), out, cache.ExpiresIn()); err != nil {
			logger.Error("refresh cache expire fail", "key", cache.Key(), "err", err)
		}
	}
	return nil
}

// SetData update database and remove cache.
func (m *Model) SetData(ctx context.Context, set func() error, caches ...Cacher) error {
	// remove cache
	for _, cache := range caches {
		if err := m.delCache(ctx, cache); err != nil {
			return err
		}
	}
	// update database
	if err := set(); err != nil {
		return err
	}

	// remove cache again
	go func(ctx context.Context, caches ...Cacher) {
		for _, cache := range caches {
			// 删除缓存
			if err := m.delCache(ctx, cache); err != nil {
				logger.L(ctx).Error("delay delete cache fail", "err", err)
			}
		}
	}(ctx, caches...)
	return nil
}

// SetAndGetData update database and update cache.
func (m *Model) SetAndGetData(ctx context.Context, out any, cache Cacher, set func() (any, error)) error {
	if out == nil {
		logger.Error("SetAndGetData out is nil")
		return status.InternalServerError()
	}

	if err := m.delCache(ctx, cache); err != nil {
		return err
	}

	result, err := set()
	if err != nil {
		return err
	}

	if cache != nil && cache.Enabled() {
		if err := cache.Set(ctx, cache.Key(), result, cache.ExpiresIn()); err != nil {
			logger.Error("SetAndGetData set cache fail", "key", cache.Key(), "err", err)
		}
	}
	return m.copy(result, out)
}

func (m *Model) delCache(ctx context.Context, cache Cacher) error {
	if cache == nil || !cache.Enabled() {
		return nil
	}
	// 删除缓存数据
	if err := cache.Del(ctx, cache.Key()); err != nil {
		logger.Error("delete cache fail", "key", cache.Key(), "err", err)
		return status.DeleteCacheFailError()
	}
	return nil
}

func (m *Model) copy(from, to any) error {
	fromVal := reflect.ValueOf(from)
	toVal := reflect.ValueOf(to)
	if toVal.Kind() != reflect.Ptr || toVal.IsNil() {
		logger.Error("out must be a non-nil ptr")
		return status.InternalServerError()
	}
	if fromVal.Type() != toVal.Type() {
		logger.Error("type mismatch: get func return type must be same with out type")
		return status.InternalServerError()
	}
	toVal.Elem().Set(fromVal.Elem())
	return nil
}
