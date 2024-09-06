package mutex

// RWLocker 读写锁需要实现的方法
type RWLocker interface {
	Locker
	RLock()
	RUnlock()
}

// RWMutex 读写锁
type RWMutex struct {
	RWLocker RWLocker
}
