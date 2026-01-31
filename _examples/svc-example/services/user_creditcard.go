package services

import (
	"svc-example/datas"
	"sync"

	"github.com/asjard/asjard/core/bootstrap"
)

type UserCreditCardSvc struct {
	datas.UserCreditCard
}

var (
	userCreditCardSvc     *UserCreditCardSvc
	userCreditCardSvcOnce sync.Once
)

func NewUserCreditCardSvc() *UserCreditCardSvc {
	userCreditCardSvcOnce.Do(func() {
		userCreditCardSvc = &UserCreditCardSvc{}
		bootstrap.AddBootstrap(userCreditCardSvc)
	})
	return userCreditCardSvc
}

func (s *UserCreditCardSvc) Start() error { return nil }
func (s *UserCreditCardSvc) Stop()        {}
