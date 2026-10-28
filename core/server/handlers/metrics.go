package handlers

import (
	"context"
	"io"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/protobuf/metricspb"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MetricsAPI metrics相关接口
type MetricsAPI struct {
	reg prometheus.TransactionalGatherer
	metricspb.UnimplementedMetricsServer
}

// NewMetricsAPI metrics接口
func NewMetricsAPI(gather prometheus.Gatherer) *MetricsAPI {
	return &MetricsAPI{reg: prometheus.ToTransactionalGatherer(gather)}
}

// Fetch 获取系统指标
func (api *MetricsAPI) Fetch(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	htx, ok := ctx.(rest.RestContext)
	if !ok {
		return &emptypb.Empty{}, nil
	}
	return api.writeMetrics(htx.Headers(), func(contentType expfmt.Format) {
		htx.SetContentType(string(contentType))
	}, htx)
}

func (api *MetricsAPI) writeMetrics(headers map[string][]string, setContentType func(expfmt.Format), writer io.Writer) (*emptypb.Empty, error) {
	mfs, done, err := api.reg.Gather()
	defer done()
	if err != nil {
		logger.Error("gathering metrics fail", "err", err)
		return nil, status.InternalServerError()
	}
	contentType := expfmt.NegotiateIncludingOpenMetrics(headers)
	setContentType(contentType)
	enc := expfmt.NewEncoder(writer, contentType)
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			logger.Error("encoding and sending metric family fail", "err", err)
			return nil, status.InternalServerError()
		}
	}
	if closer, ok := enc.(expfmt.Closer); ok {
		if err := closer.Close(); err != nil {
			logger.Error("closer close fail", "err", err)
			return nil, status.InternalServerError()
		}
	}
	return nil, nil
}

func (MetricsAPI) RestServiceDesc() *rest.ServiceDesc {
	return &metricspb.MetricsRestServiceDesc
}
