package main

import (
	"log"

	configpb "protos-repo/example/api/v1/config"
	"protos-repo/example/api/v1/hello"
	pb "protos-repo/example/api/v1/sample"
	"protos-repo/example/api/v1/user"
	apiv1 "svc-gw/apis/api/v1"
	_ "svc-gw/interceptors"

	"github.com/asjard/asjard/v2"
	"github.com/asjard/asjard/v2/pkg/server/mcp"
	"github.com/asjard/asjard/v2/pkg/server/rest"
)

func main() {
	server := asjard.New()

	if err := server.AddHandlers(rest.Protocol,
		&user.UserAPI{},
		&pb.SampleAPI{},
		&hello.HelloAPI{},
		&configpb.ConfigAPI{},
		&apiv1.GwAPI{}); err != nil {
		log.Fatal(err)
	}

	if err := server.AddHandlers(mcp.Protocol,
		&hello.HelloMcpAPI{},
		&user.UserMcpAPI{},
		&pb.SampleMcpAPI{}); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Start())
}
