package main

import (
	"log"

	apiv1 "svc-example/apis/api/v1"
	openapiv1 "svc-example/apis/openapi/v1"
	"svc-example/services"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

func main() {
	server := asjard.New()

	svcCtx := services.NewServiceContext()

	if err := server.AddHandler(openapiv1.NewUserAPI(svcCtx), grpc.Protocol, rest.Protocol); err != nil {
		log.Fatal(err)
	}

	if err := server.AddHandler(apiv1.NewUserAPI(svcCtx), grpc.Protocol, rest.Protocol); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Start())
}
