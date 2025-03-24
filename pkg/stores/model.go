package stores

import (
	"context"
	"reflect"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"golang.org/x/sync/singleflight"
)

// Model 带缓存的数据存储，获取，删除
type Model struct {
	sg singleflight.Group
}

// Modeler 模型需要实现的方法
type Modeler interface {
	// 返回库名和表名
	ModelName() string
}

const (
	// DefaultSingleflightKey 默认singleflight
	DefaultSingleflightKey = "default"
)

// GetData 获取数据
// 先从缓存获取数据，如果缓存中没有数据则从数据源获取数据
// out: 将获取到的数据赋值给这个参数
// cacher: 将从这个缓存获取数据，如果获取不到数据则会执行do方法
// get: 从数据源获取数据
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
	// 从缓存获取数据
	fromCurrent, err := cache.Get(ctx, cache.Key(), out)
	if err != nil {
		// 从数据源获取数据
		result, err, _ := m.sg.Do(cache.Key(), get)
		if err != nil {
			// 如果从数据源也没有获取到数据
			// 则可以设置一个空值
			if rerr := cache.Set(ctx, cache.Key(), out, cache.EmptyExpiresIn()); rerr != nil {
				logger.Error("set empty into cache fail", "err", rerr)
			}
			return err
		}
		// 设置缓存
		if err := cache.Set(ctx, cache.Key(), result, cache.ExpiresIn()); err != nil {
			logger.Error("set cache fail", "key", cache.Key(), "err", err)
		}
		return m.copy(result, out)
	}
	// 刷新缓存时间
	// 如果获取到的数据是从当前缓存中获取到的则刷新缓存
	if fromCurrent && cache.AutoRefresh() {
		if err := cache.Refresh(ctx, cache.Key(), out, cache.ExpiresIn()); err != nil {
			logger.Error("refresh cache expire fail", "key", cache.Key(), "err", err)
		}
	}
	return nil
}

// SetData 更新数据源并删除缓存
// 先更新数据源，然后删除缓存
// set: 更新数据源数据, 如果删除缓存过程中出现失败则会通过管道通知
// caches: 缓存
func (m *Model) SetData(ctx context.Context, cache Cacher, set func() error) error {
	// 更新数据源数据
	if err := set(); err != nil {
		return err
	}
	// 删除缓存
	if err := m.delCache(ctx, cache); err != nil {
		return err
	}
	return nil
}

// SetAndGetData 更新数据并返回更新后的数据
func (m *Model) SetAndGetData(ctx context.Context, out any, cache Cacher, set func() (any, error)) error {
	if out == nil {
		logger.Error("SetAndGetData out is nil")
		return status.InternalServerError()
	}
	result, err := set()
	if err != nil {
		return err
	}

	if err := m.delCache(ctx, cache); err != nil {
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

// from值拷贝到to
// from和to必须是同类型
// from 和to 必须是指针
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
