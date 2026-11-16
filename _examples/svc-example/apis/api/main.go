package main

import (
	"log"

	apiv1 "svc-example/apis/api/v1"
	"svc-example/services"

	"github.com/asjard/asjard/v2"
	"github.com/asjard/asjard/v2/pkg/server/grpc"
	"github.com/asjard/asjard/v2/pkg/server/rest"
)

func main() {
	server := asjard.New()

	svcCtx := services.NewServiceContext()

	if err := server.AddHandlers(grpc.Protocol,
		apiv1.NewUserAPI(svcCtx),
		apiv1.NewHelloAPI(),
		&apiv1.ConfigAPI{},
		&apiv1.SampleAPI{}); err != nil {
		log.Fatal(err)
	}

	if err := server.AddHandlers(rest.Protocol,
		&apiv1.ConfigAPI{},
		&apiv1.SampleAPI{}); err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Start())
}
