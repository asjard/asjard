package services

import (
	"context"

	"svc-example/datas"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/pkg/stores/xgorm"
)

type Svcs struct {
	UserSvc           *UserSvc
	UserCreditCardSvc *datas.UserCreditCard
}
type ServiceContext struct {
	*Svcs
}

func NewServiceContext() *ServiceContext {
	s := &ServiceContext{}
	bootstrap.AddBootstrap(s)
	s.Svcs = &Svcs{
		UserSvc:           NewUserSvc(),
		UserCreditCardSvc: &datas.UserCreditCard{},
	}
	return s
}

func (s *ServiceContext) Start() error {
	if config.GetBool("dbAutoMigrate", false) {
		db, err := xgorm.DB(context.Background())
		if err != nil {
			return err
		}
		if err := db.AutoMigrate(&datas.User{}, &datas.UserCreditCard{}); err != nil {
			return err
		}
	}
	return nil
}
func (s *ServiceContext) Stop() {}
