package apiv1

import (
	"context"

	pb "github.com/asjard/asjard/_examples/protos-repo/gw/api/v1/gw"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/pkg/server/rest"
)

type GwAPI struct {
	pb.UnimplementedGwServer
}

func (api *GwAPI) Start() error                       { return nil }
func (api *GwAPI) Stop()                              {}
func (api *GwAPI) RestServiceDesc() *rest.ServiceDesc { return &pb.GwRestServiceDesc }

func (api *GwAPI) GetServiceInstances(ctx context.Context, in *pb.ServiceInstancesReq) (*pb.ServiceInstancesResp, error) {
	var instances []*pb.ServiceInstancesResp_Instance
	for _, item := range registry.PickServices(registry.WithServiceName(in.ServiceName)) {
		instances = append(instances, &pb.ServiceInstancesResp_Instance{
			ServiceName: item.Service.Instance.Name,
			InstanceId:  item.Service.Instance.ID,
		})
	}

	return &pb.ServiceInstancesResp{
		Instances: instances,
	}, nil
}
