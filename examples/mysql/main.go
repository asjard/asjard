package main

import (
	"log"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/examples/mysql/handler"
	"github.com/asjard/asjard/pkg/server/rest"
)

func main() {
	server := asjard.New()
	mysqlAPI := handler.NewMysqlExampleAPI()
	server.AddHandler(rest.Protocol, mysqlAPI)
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
