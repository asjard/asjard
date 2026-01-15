package mutex

// RWLocker defines the contract for a distributed read-write lock implementation.
// It extends the basic Locker interface, adding specialized methods for
// shared (read) access.
type RWLocker interface {
	// Locker provides the standard Lock, Unlock, and KeepAlive methods for exclusive writing.
	Locker

	// RLock attempts to acquire a shared read lock.
	// Multiple callers can hold a read lock simultaneously, provided no exclusive
	// write lock is active.
	RLock()

	// RUnlock releases a previously acquired shared read lock.
	RUnlock()
}

// RWMutex is a high-level wrapper that provides the distributed read-write locking logic.
// It uses an underlying RWLocker implementation (e.g., backed by Redis or ETCD).
type RWMutex struct {
	// RWLocker is the concrete implementation of the distributed locking primitives.
	RWLocker RWLocker
}
