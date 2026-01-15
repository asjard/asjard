package xgorm

import (
	"context"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/mutex"
)

// Lock represents the database schema for the distributed lock.
// It uses a unique index on 'lock_key' to ensure that only one owner can hold the lock at a time.
type Lock struct {
	Id int64 `gorm:"column:id;type:BIGINT(20);primaryKey;autoIncrement;comment:主键"`
	// LockKey is the unique identifier for the resource being locked.
	LockKey string `gorm:"column:lock_key;type:VARCHAR(255);uniqueIndex;comment:锁名称"`
	// Owner (threadId) identifies who currently holds the lock, preventing unauthorized unlocks.
	Owner     string `gorm:"column:owner;type:VARCHAR(36);comment:当前锁拥有者"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// ExpiresAt handles deadlocks by allowing locks to be cleaned up if the owner crashes.
	ExpiresAt time.Time

	// options stores GORM connection configurations (e.g., which database instance to use).
	options []Option `gorm:"-"`
}

// Verify that Lock implements the mutex.Locker interface.
var _ mutex.Locker = &Lock{}

// NewLock initializes a new GORM-based locker and starts a background cleanup routine.
func NewLock(opts ...Option) (mutex.Locker, error) {
	lock := &Lock{
		options: opts,
	}
	// Start a background goroutine to clean up expired locks.
	go lock.cleanUp()
	return lock, nil
}

// Lock attempts to acquire the lock by inserting a record into the database.
// Returns true if acquired, false if the key already exists (lock held by someone else) or on error.
func (l Lock) Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	db, err := DB(ctx, l.options...)
	if err != nil {
		logger.Error("gorm lock get db fail", "key", key, "thread_id", threadId, "err", err)
		return false
	}
	// Create uses the unique index on LockKey. If the key exists, this will fail.
	if err := db.Model(&Lock{}).Create(&Lock{
		LockKey:   key,
		Owner:     threadId,
		ExpiresAt: time.Now().Add(expiresIn),
	}).Error; err != nil {
		return false
	}
	return true
}

// Unlock releases the lock by deleting the record.
// It ensures that only the original owner (threadId) can release the lock.
func (l Lock) Unlock(ctx context.Context, key, threadId string) bool {
	db, err := DB(ctx, l.options...)
	if err != nil {
		logger.Error("gorm unlock get db fail", "key", key, "thread_id", threadId, "err", err)
		return false
	}
	if err := db.Where("lock_key=?", key).
		Where("owner=?", threadId).
		Delete(&Lock{}).Error; err != nil {
		logger.Error("gorm unlock delete fail", "key", key, "thread_id", threadId, "err", err)
		return false
	}
	return true
}

// KeepAlive extends the expiration time of an existing lock.
// This is used for long-running tasks to prevent the lock from expiring while still in use.
func (l Lock) KeepAlive(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	db, err := DB(ctx, l.options...)
	if err != nil {
		logger.Error("gorm lock keepalive get db fail", "key", key, "thread_id", threadId, "err", err)
		return false
	}
	if err := db.Model(&Lock{}).
		Where("lock_key=?", key).
		Where("owner=?", threadId).
		Update("expires_at", time.Now().Add(expiresIn)).Error; err != nil {
		logger.Error("gorm lock keepalive refresh fail", "key", key, "thread_id", threadId, "err", err)
		return false
	}
	return true
}

// cleanUp periodically removes expired records from the database to prevent the table from growing
// indefinitely and to allow abandoned locks to be re-acquired.
func (l Lock) cleanUp() {
	// Note: This logic currently only runs once after a 1-second delay.
	// In a production scenario, this would usually be inside a for { select { case <-ticker.C: ... } } loop.
	select {
	case <-time.After(time.Second):
		db, err := DB(context.Background(), l.options...)
		if err == nil {
			if err := db.Where("expires_at<?", time.Now()).
				Delete(&Lock{}).Error; err != nil {
				logger.Error("gorm lock clean up fail", "err", err)
			}
		} else {
			logger.Error("gorm lock clean up get db fail", "err", err)
		}
	}
}
