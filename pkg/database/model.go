package database

import (
	"errors"

	"golang.org/x/sync/singleflight"
)

// Model 带缓存的数据存储，获取，删除
type Model struct{}

// 防止缓存穿透
var sg singleflight.Group

const (
	// DefaultSingleflightKey 默认singleflight
	DefaultSingleflightKey = "default"
)

// GetData 获取数据
// 先从缓存获取数据，如果缓存中没有数据则从数据源获取数据
// out: 将获取到的数据赋值个这个参数
// cache: 将从这个缓存获取数据，如果获取不到数据则会执行do方法
// do: 从数据源获取数据
// hitted: 是否命中了缓存
func (m Model) GetData(out interface{}, cache *Cache, do func() (interface{}, error)) (hitted bool, err error) {
	if out == nil {
		return false, errors.New("out is empty")
	}
	if cache == nil {
		out, err = m.doGet(DefaultSingleflightKey, do)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	// 从缓存获取数据
	if err := cache.get(out); err != nil {
		// 从数据源获取数据
		out, err = m.doGet(cache.key, do)
		if err != nil {
			// 如果从数据源也没有获取到数据
			// 则可以设置一个空值
			cache.setEmpty(out)
			return false, err
		}
		// 设置缓存
		if err := cache.set(out); err != nil {
			return false, err
		}
		return false, nil
	}
	// 刷新缓存过期时间
	return true, cache.refreshExpire()
}

// SetData 创建、更新、或删除数据
// 先更新数据源，然后删除缓存
// do: 更新数据源数据, 如果删除缓存过程中出现失败则会通过管道通知
// caches: 缓存列表
func (m Model) SetData(do func(chan error) error, caches ...*Cache) error {
	delErr := make(chan error)
	defer close(delErr)
	// 更新数据源数据
	if err := do(delErr); err != nil {
		return err
	}
	// 删除缓存数据
	for _, cache := range caches {
		if err := cache.del(); err != nil {
			delErr <- err
			return err
		}
	}
	return nil
}

func (m Model) doGet(key string, do func() (interface{}, error)) (interface{}, error) {
	v, err, _ := sg.Do(key, do)
	return v, err
}
