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
	"github.com/asjard/asjard/pkg/protobuf/requestpb"
	"github.com/asjard/asjard/pkg/stores/xgorm"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type User struct {
	Id        int64  `gorm:"type:BIGINT(20);primayKey"`
	Username  string `gorm:"type:VARCHAR(50);uniqueIndex;NOT NULL;comment:username"`
	Age       int32  `gorm:"type:INT;default:0"`
	CardNums  int32  `gorm:"type:INT;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t User) TableName() string { return "user" }
func (t User) ModelName() string { return t.TableName() }

func (t User) Create(ctx context.Context, in *user.UserReq) error {
	db, err := xgorm.DB(ctx)
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

func (t User) Update(ctx context.Context, in *user.UserReq) error {
	db, err := xgorm.DB(ctx)
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
	db, err := xgorm.DB(ctx)
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

func (t User) Get(ctx context.Context, in *cpb.ReqWithName) (*user.UserInfo, error) {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return nil, err
	}
	var record User
	if err := db.Where("username=?", in.Name).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Code(xcodes.ERR_USER_EUSE_NOT_FOUND), "user '%s' not found", in.Name)
		}
		logger.L(ctx).Error("get user fail", "username", in.Name, "err", err)
		return nil, status.InternalServerError()
	}

	return record.Info(), nil
}

func (t User) Del(ctx context.Context, in *cpb.ReqWithName) error {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return err
	}
	if err := db.Where("username=?", in.Name).Delete(&User{}).Error; err != nil {
		logger.L(ctx).Error("delete user fail", "username", in.Name, "err", err)
		return status.InternalServerError()
	}
	return nil
}
func (t User) Search(ctx context.Context, in *user.UserSearchReq) (*user.UserList, error) {
	db, err := xgorm.DB(ctx)
	if err != nil {
		return nil, err
	}
	var records []User
	var total int64
	if err := db.Model(&User{}).Scopes(t.searchFilter(in)).
		Count(&total).
		Scopes(requestpb.ReqWithPageGormScope(in.Page, in.Size, in.Sort, "created_at")).
		Find(&records).Error; err != nil {
		logger.L(ctx).Error("search user fail", "req", in, "err", err)
		return nil, status.InternalServerError()
	}
	list := make([]*user.UserInfo, len(records))
	for idx, record := range records {
		list[idx] = record.Info()
	}
	return &user.UserList{
		Total: int32(total),
		List:  list,
	}, nil
}

func (t User) searchFilter(in *user.UserSearchReq) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if in.Keywords != "" {
			db = db.Where("username like ?", in.Keywords+"%")
		}
		return db
	}
}

func (t User) Info() *user.UserInfo {
	return &user.UserInfo{
		UserId:   t.Id,
		Username: t.Username,
		Age:      t.Age,
		CardNums: t.CardNums,
	}
}
