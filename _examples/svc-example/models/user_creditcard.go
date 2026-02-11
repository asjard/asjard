package models

import (
	"svc-example/datas"
	"sync"

	"github.com/asjard/asjard/core/bootstrap"
)

type UserCreditCardModel struct {
	datas.UserCreditCard
}

var (
	userCreditCardModel     *UserCreditCardModel
	userCreditCardModelOnce sync.Once
)

func NewUserCreditCardModel() *UserCreditCardModel {
	userCreditCardModelOnce.Do(func() {
		userCreditCardModel = &UserCreditCardModel{}
		bootstrap.AddBootstrap(userCreditCardModel)
	})
	return userCreditCardModel
}

func (s *UserCreditCardModel) Start() error { return nil }
func (s *UserCreditCardModel) Stop()        {}
