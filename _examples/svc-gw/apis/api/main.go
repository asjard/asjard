package main

import (
	"log"

	configpb "protos-repo/example/api/v1/config"
	pb "protos-repo/example/api/v1/sample"
	apiv1 "svc-gw/apis/api/v1"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/pkg/server/rest"
)

func main() {
	server := asjard.New()

	if err := server.AddHandlers(rest.Protocol,
		&pb.SampleAPI{},
		&configpb.ConfigAPI{},
		&apiv1.GwAPI{}); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Start())
}
