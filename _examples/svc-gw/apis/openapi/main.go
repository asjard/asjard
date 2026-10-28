package main

import (
	"log"

	"protos-repo/example/openapi/v1/user"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/pkg/server/mcp"
	"github.com/asjard/asjard/pkg/server/rest"
	_ "github.com/asjard/asjard/pkg/server/rest/xfasthttp"
)

func main() {
	server := asjard.New()

	if err := server.AddHandler(&user.UserAPI{}, rest.Protocol, mcp.Protocol); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Start())
}
