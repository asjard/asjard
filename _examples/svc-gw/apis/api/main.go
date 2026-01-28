package main

import (
	"log"

	pb "github.com/asjard/asjard/_examples/protos-repo/example/api/v1/sample"
	apiv1 "github.com/asjard/asjard/_examples/svc-gw/apis/api/v1"
	"github.com/asjard/asjard/pkg/server/rest"

	"github.com/asjard/asjard"
)

func main() {
	server := asjard.New()

	if err := server.AddHandlers(rest.Protocol, &pb.SampleAPI{}, &apiv1.GwAPI{}); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Start())
}
