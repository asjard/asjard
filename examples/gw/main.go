package main

import (
	"github.com/asjard/asjard"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	"github.com/asjard/asjard/pkg/server/rest"
)

var _ pb.HelloServer = &pb.HelloAPI{}
var _ rest.Handler = &pb.HelloAPI{}

func main() {
	server := asjard.New()
	server.AddHandler(&pb.HelloAPI{}, rest.Protocol)
	if err := server.Start(); err != nil {
		panic(err)
	}
}
