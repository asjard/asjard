package datas

import (
	"context"
	"errors"
	"svc-example/datas/models"

	"protos-repo/common/xcodes"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/protobuf/requestpb"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type User struct {
	Model

	Username string `gorm:"type:VARCHAR(50);uniqueIndex;NOT NULL;comment:username"`
	Age      int32  `gorm:"type:INT;default:0"`
	CardNums int32  `gorm:"type:INT;default:0"`
}

func (t User) TableName() string { return "user" }

func (t User) Create(ctx context.Context, in *User) error {
	db, err := t.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Create(&User{
		Username: in.Username,
		Age:      in.Age,
	}).Error; err != nil {
		logger.L(ctx).Error("create user fail", "req", in, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t User) Update(ctx context.Context, in *User) error {
	db, err := t.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Model(&User{}).Where("username=?", in.Username).
		Updates(map[string]any{
			"age": in.Age,
		}).Error; err != nil {
		logger.L(ctx).Error("update user fail", "req", in, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t User) UpdateCardNum(ctx context.Context, username string, num int) error {
	db, err := t.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Model(&User{}).
		Where("username=?", username).
		Update("card_nums", gorm.Expr("card_nums+?", num)).Error; err != nil {
		logger.L(ctx).Error("update user card nums fail", "username", username, "err", err)
		return status.InternalServerError()
	}
	return nil
}

func (t User) Get(ctx context.Context, username string) (*User, error) {
	db, err := t.DB(ctx)
	if err != nil {
		return nil, err
	}
	var record User
	if err := db.Where("username=?", username).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Code(xcodes.ERR_USER_EUSE_NOT_FOUND), "user '%s' not found", username)
		}
		logger.L(ctx).Error("get user fail", "username", username, "err", err)
		return nil, status.InternalServerError()
	}

	return &record, nil
}

// Del user.
func (t User) Del(ctx context.Context, username string) error {
	db, err := t.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Where("username=?", username).Delete(&User{}).Error; err != nil {
		logger.L(ctx).Error("delete user fail", "username", username, "err", err)
		return status.InternalServerError()
	}
	return nil
}

// Search users.
func (t User) Search(ctx context.Context, in *models.UserSearchReq) (int32, []*User, error) {
	db, err := t.DB(ctx)
	if err != nil {
		return 0, nil, err
	}
	var records []*User
	var total int64
	if err := db.Model(&User{}).Scopes(t.searchFilter(in)).
		Count(&total).
		Scopes(requestpb.ReqWithPageGormScope(in.Page, in.Size, in.Sort, "created_at")).
		Find(&records).Error; err != nil {
		logger.L(ctx).Error("search user fail", "req", in, "err", err)
		return 0, nil, status.InternalServerError()
	}
	return int32(total), records, nil
}

func (t User) searchFilter(in *models.UserSearchReq) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if in.Keywords != "" {
			db = db.Where("username like ?", in.Keywords+"%")
		}
		return db
	}
}
