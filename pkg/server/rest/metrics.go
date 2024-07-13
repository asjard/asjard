package rest

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type MetricsAPI struct {
	UnimplementedMetricsServer
}

// Fetch 获取系统指标
func (MetricsAPI) Fetch(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (MetricsAPI) RestServiceDesc() *ServiceDesc {
	return &MetricsRestServiceDesc
}
