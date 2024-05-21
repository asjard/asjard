package database

// RedisCache redis缓存
type RedisCache struct{}

var _ Cacher = &RedisCache{}

func (c RedisCache) get(out interface{}) error {
	return nil
}

func (c RedisCache) del() error {
	return nil
}

func (c RedisCache) set(in interface{}) error {
	return nil
}

func (c RedisCache) refreshExpire() error {
	return nil
}

func (c RedisCache) setEmpty(in interface{}) error {
	return nil
}
