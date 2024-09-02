package xgorm

import (
	"context"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/mutex"
)

// Lock 数据库锁
type Lock struct {
	Id        int64  `gorm:"column:id;type:BIGINT(20);primaryKey;autoIncrement;comment:主键"`
	LockKey   string `gorm:"column:lock_key;type:VARCHAR(255);uniqueIndex;comment:锁名称"`
	Owner     string `gorm:"column:owner;type:VARCHAR(36);comment:当前锁拥有者"`
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time

	options []Option `gorm:"-"`
}

var _ mutex.Locker = &Lock{}

func NewLock(opts ...Option) (mutex.Locker, error) {
	lock := &Lock{
		options: opts,
	}
	go lock.cleanUp()
	return lock, nil
}

// Lock 加锁
func (l Lock) Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool {
	db, err := DB(ctx, l.options...)
	if err != nil {
		logger.Error("gorm lock get db fail", "key", key, "thread_id", threadId, "err", err)
		return false
	}
	if err := db.Model(&Lock{}).Create(&Lock{
		LockKey:   key,
		Owner:     threadId,
		ExpiresAt: time.Now().Add(expiresIn),
	}).Error; err != nil {
		return false
	}
	return true
}

// Unlock 解锁
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

// 续期
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

func (l Lock) cleanUp() {
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
