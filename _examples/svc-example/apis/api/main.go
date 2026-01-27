package main

import (
	"log"

	"github.com/asjard/asjard"
	apiv1 "github.com/asjard/asjard/_examples/svc-example/apis/api/v1"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

func main() {
	server := asjard.New()
	if err := server.AddHandler(apiv1.NewSampleAPI(), grpc.Protocol, rest.Protocol); err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Start())
}
