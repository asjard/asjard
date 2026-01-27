package main

import (
	"log"

	pb "github.com/asjard/asjard/_examples/protos-repo/example/api/v1/sample"
	"github.com/asjard/asjard/pkg/server/rest"

	"github.com/asjard/asjard"
)

func main() {
	server := asjard.New()

	if err := server.AddHandler(&pb.SampleAPI{}, rest.Protocol); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Start())
}
