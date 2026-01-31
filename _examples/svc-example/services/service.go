package services

import (
	"context"
	"svc-example/datas"
	"sync"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/pkg/stores/xgorm"
)

type Svcs struct {
	UserSvc           *UserSvc
	UserCreditCardSvc *UserCreditCardSvc
}
type ServiceContext struct {
	Svcs *Svcs
}

var (
	serviceContext     *ServiceContext
	serviceContextOnce sync.Once
)

func NewServiceContext() *ServiceContext {
	serviceContextOnce.Do(func() {
		serviceContext = &ServiceContext{}
		bootstrap.AddBootstrap(serviceContext)

		serviceContext.Svcs = &Svcs{
			UserSvc:           NewUserSvc(),
			UserCreditCardSvc: NewUserCreditCardSvc(),
		}
	})
	return serviceContext
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
