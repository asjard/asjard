package main

import (
	"log"

	"protos-repo/example/openapi/v1/user"

	"github.com/asjard/asjard/v2"
	"github.com/asjard/asjard/v2/pkg/server/rest"
)

func main() {
	server := asjard.New()

	if err := server.AddHandlers(rest.Protocol, &user.UserAPI{}); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Start())
}
