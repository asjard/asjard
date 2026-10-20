package datas

import (
	"context"
	"time"

	"github.com/asjard/asjard/pkg/stores/xgorm"
	"gorm.io/gorm"
)

type Model struct {
	ID        int64     `gorm:"type:BIGINT(20);primarykey"`
	CreatedAt time.Time `gorm:"index;comment:创建时间"`
	UpdatedAt time.Time
}

type ctxKey int

const (
	isTransaction ctxKey = 1
)

func (Model) DB(ctx context.Context) (*gorm.DB, error) { return xgorm.DB(ctx) }

func (t Model) DoWithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 如果已经开启了事物
	if _, ok := ctx.Value(isTransaction).(bool); ok {
		return fn(ctx)
	}
	db, err := t.DB(ctx)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(xgorm.WithDB(ctx, tx), isTransaction, true))
	})
}
