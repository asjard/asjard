package services

import (
	"context"
	"sync"

	"svc-example/datas"
	"svc-example/models"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/pkg/stores/xgorm"
)

type Models struct {
	UserModel           *models.UserModel
	UserCreditCardModel *models.UserCreditCardModel
}
type ServiceContext struct {
	Models *Models
}

var (
	serviceContext     *ServiceContext
	serviceContextOnce sync.Once
)

func NewServiceContext() *ServiceContext {
	serviceContextOnce.Do(func() {
		serviceContext = &ServiceContext{}
		bootstrap.AddBootstrap(serviceContext)
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
	s.Models = &Models{
		UserModel:           models.NewUserModel(),
		UserCreditCardModel: models.NewUserCreditCardModel(),
	}
	return nil
}
func (s *ServiceContext) Stop() {}
