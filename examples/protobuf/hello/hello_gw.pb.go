package hello

// import (
// 	"context"

// 	"github.com/asjard/asjard/core/client"
// 	mgrpc "github.com/asjard/asjard/pkg/client/grpc"
// 	"github.com/asjard/asjard/pkg/server/rest"
// )

// type HelloAPI struct {
// 	UnimplementedHelloServer
// 	client HelloClient
// }

// func (api *HelloAPI) Bootstrap() error {
// 	conn, err := client.NewClient(mgrpc.Protocol, "HelloGrpc").Conn()
// 	if err != nil {
// 		return err
// 	}
// 	api.client = NewHelloClient(conn)
// 	return nil
// }

// func (api *HelloAPI) Shutdown() {}

// func (api *HelloAPI) Say(ctx context.Context, in *SayReq) (*SayReq, error) {
// 	return api.client.Say(ctx, in)
// }

// func (api *HelloAPI) RestServiceDesc() *rest.ServiceDesc {
// 	return &HelloRestServiceDesc
// }
