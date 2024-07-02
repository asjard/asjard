package mysql

// import (
// 	"time"
// )

// // Lock 数据库锁
// type Lock struct {
// 	Id        int64  `gorm:"column:id;type:INT(20);primaryKey;autoIncrement;comment:主键"`
// 	LockKey   string `gorm:"column:lock_key;type:VARCHAR(255);uniqueIndex;comment:锁名称"`
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// 	options   *lockOptions `gorm:"-"`
// }

// type lockOptions struct {
// 	timeout time.Duration
// 	lockKey string
// }

// func NewLock(lockKey string, timeout time.Duration) *Lock {
// 	return &Lock{
// 		options: &lockOptions{
// 			timeout: timeout,
// 			lockKey: lockKey,
// 		},
// 	}
// }

// // Lock 加锁
// func (Lock) Lock(ctx context.Context, options ...Option) (bool, error) {
// 	db, err := DB(ctx, options...)
// 	if err != nil {
// 		return false, err
// 	}
// 	return false, nil
// }

// // Unlock 解锁
// func (Lock) Unlock(ctx context.Context, options ...Option) error {
// 	db, err := DB(ctx, options...)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // TryLock 尝试加锁
// func (Lock) TryLock(timeout time.Duration) error {
// 	return nil
// }
