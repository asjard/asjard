package apiv1

import (
	"context"

	cpb "protos-repo/common/common"
	pb "protos-repo/example/api/v1/config"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
)

type ConfigAPI struct {
	pb.UnimplementedConfigServer
}

func (api *ConfigAPI) Start() error {
	return config.Set("example.config.configs.key_in_different_sourcer", "in_mem_source")
}
func (api *ConfigAPI) Stop()                              {}
func (api *ConfigAPI) GrpcServiceDesc() *grpc.ServiceDesc { return &pb.Config_ServiceDesc }
func (api *ConfigAPI) RestServiceDesc() *rest.ServiceDesc { return &pb.ConfigRestServiceDesc }

// Simple Get: Retrieves the current configuration.
// Asjard generates both a gRPC method and a REST endpoint.
func (api *ConfigAPI) Get(ctx context.Context, in *pb.ConfigGetReq) (*pb.ConfigGetResp, error) {

	app := runtime.GetAPP()
	result := pb.ConfigGetResp{
		Page: in.Page,
		Size: in.Size,
		Sort: in.Sort,
		Instance: &pb.ConfigGetResp_Instance{
			Id:         app.Instance.ID,
			Name:       app.Instance.Name,
			SystemCode: app.Instance.SystemCode,
			Version:    app.Instance.Version,
			Metadata:   app.Instance.MetaData,
		},
		App:    app.App,
		Region: app.Region,
		Az:     app.AZ,
	}
	if err := config.GetWithJsonUnmarshal("example.config", &result); err != nil {
		logger.L(ctx).Error("get config with json unmarshal fail", "err", err)
		return nil, status.InternalServerError()
	}
	return &result, nil
}

// GetAndDecrypt: Retrieves an encrypted configuration and returns it in plain text.
// Demonstrates handling specialized business logic through the same framework.
func (api *ConfigAPI) GetAndDecrypt(ctx context.Context, in *cpb.Empty) (*pb.ConfigDecryptResp, error) {
	return &pb.ConfigDecryptResp{
		PlainText: config.GetString("example.config.encrypted_value", ""),
	}, nil
}
