package main

import (
	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/examples/mysql/model"
)

type Bootstrap struct{}

func init() {
	bootstrap.AddBootstrap(&Bootstrap{})
}

func (Bootstrap) Bootstrap() error {
	// 数据库初始化
	if err := model.Init(); err != nil {
		return err
	}
	return nil
}

func (Bootstrap) Shutdown() {}
