package datas

import (
	"context"
	"errors"
	"time"

	cpb "protos-repo/common/common"
	"protos-repo/common/xcodes"
	"protos-repo/example/api/v1/user"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/stores/xgorm"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type UserCreditCard struct {
	Id        int64 `gorm:"type:BIGINT(20);primayKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Username string `gorm:"type:VARCHAR(50);index;uniqueIndex:user_credit_card"`
	Number   string `gorm:"type:VARCHAR(100);index;uniqueIndex:user_credit_card"`
}

func (t *UserCreditCard) TableName() string { return "user_creditcard" }
func (t *UserCreditCard) ModelName() string { return t.TableName() }

func (t *UserCreditCard) Add(ctx context.Context, in *user.UserCreditCardReq) error {
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

func (t *UserCreditCard) Remove(ctx context.Context, in *user.UserCreditCardReq) error {
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

func (t *UserCreditCard) RemoveByUser(ctx context.Context, in *cpb.ReqWithName) error {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Where("username=?", in.Name).Delete(&UserCreditCard{}).Error; err != nil {
		logger.L(ctx).Error("remove user cards fail", "req", in, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t *UserCreditCard) Get(ctx context.Context, in *user.UserCreditCardReq) (*user.UserCreditCardInfo, error) {
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
	return record.Info(), nil
}

func (t *UserCreditCard) Search(ctx context.Context, in *cpb.ReqWithName) (*user.UserCreditCardList, error) {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return nil, err
	}
	var records []UserCreditCard
	var total int64
	if err := db.Model(&UserCreditCard{}).Where("username=?", in.Name).
		Count(&total).
		Find(&records).Error; err != nil {
		logger.L(ctx).Error("search user credit card fail", "req", in, "err", err)
		return nil, status.InternalServerError()
	}
	list := make([]*user.UserCreditCardInfo, len(records))
	for idx, item := range records {
		list[idx] = item.Info()
	}
	return &user.UserCreditCardList{
		Total: int32(total),
		List:  list,
	}, nil
}

func (t *UserCreditCard) Info() *user.UserCreditCardInfo {
	return &user.UserCreditCardInfo{
		Number: t.Number,
	}
}
