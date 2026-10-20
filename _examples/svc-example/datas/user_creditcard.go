package datas

import (
	"context"
	"errors"

	"protos-repo/common/xcodes"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/stores/xgorm"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type UserCreditCard struct {
	Model

	Username string `gorm:"type:VARCHAR(50);index;uniqueIndex:user_credit_card"`
	Number   string `gorm:"type:VARCHAR(100);index;uniqueIndex:user_credit_card"`
}

func (t *UserCreditCard) TableName() string { return "user_creditcard" }

func (t *UserCreditCard) Add(ctx context.Context, in *UserCreditCard) error {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Create(&UserCreditCard{
		Username: in.Username,
		Number:   in.Number,
	}).Error; err != nil {
		logger.L(ctx).Error("user add credit card fail", "req", in, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t *UserCreditCard) Remove(ctx context.Context, in *UserCreditCard) error {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Where("username=?", in.Username).
		Where("number=?", in.Number).
		Delete(&UserCreditCard{}).Error; err != nil {
		logger.L(ctx).Error("remove user card fail", "req", in, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t *UserCreditCard) RemoveByUser(ctx context.Context, username string) error {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Where("username=?", username).Delete(&UserCreditCard{}).Error; err != nil {
		logger.L(ctx).Error("remove user cards fail", "username", username, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t *UserCreditCard) Get(ctx context.Context, in *UserCreditCard) (*UserCreditCard, error) {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return nil, err
	}
	var record UserCreditCard
	if err := db.Where("username=?", in.Username).
		Where("number=?", in.Number).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Code(xcodes.ERR_USER_EUSE_CREDIT_CARD_NOT_FOUND),
				"card '%s' in user '%s' not found", in.Number, in.Username)
		}
		logger.L(ctx).Error("get user credit card fail", "req", in, "err", err)
		return nil, status.InternalServerError()
	}
	return &record, nil
}

func (t *UserCreditCard) Search(ctx context.Context, username string) (int32, []*UserCreditCard, error) {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return 0, nil, err
	}
	var records []*UserCreditCard
	var total int64
	if err := db.Model(&UserCreditCard{}).Where("username=?", username).
		Count(&total).
		Find(&records).Error; err != nil {
		logger.L(ctx).Error("search user credit card fail", "username", username, "err", err)
		return 0, nil, status.InternalServerError()
	}
	return int32(total), records, nil
}
