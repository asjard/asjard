package main

import (
	"log"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/examples/mysql/apis"
	"github.com/asjard/asjard/examples/mysql/model"
	"github.com/asjard/asjard/pkg/server/rest"
)

type System struct{}

// Bootstrap 系统初始化
func (System) Bootstrap() error {
	if err := model.Init(); err != nil {
		return err
	}
	return nil
}

// Shutdown 系统停止
func (System) Shutdown() {}

func init() {
	bootstrap.AddBootstrap(&System{})
}

func main() {
	server := asjard.New()
	server.AddHandler(apis.NewExampleAPI(), rest.Protocol)
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
