package main

import (
	"log"

	"github.com/asjard/asjard"
	hpb "github.com/asjard/asjard/examples/protobuf/hello"
	"github.com/asjard/asjard/pkg/server/rest"
)

// Hello .
type Hello struct{}

// Say .
func (Hello) Say(ctx *rest.Context, in *hpb.Say) (*hpb.Say, error) {
	return in, nil
}

// ServiceDesc .
func (Hello) ServiceDesc() rest.ServiceDesc {
	return hpb.HelloRestServiceDesc
}

func main() {
	server := asjard.New()
	server.AddHandler(rest.Protocol, &Hello{})
	if err := server.Start(); err != nil {
		log.Println(err.Error())
	}
}
