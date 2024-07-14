package handlers

import (
	"context"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/ajerr"
	"github.com/asjard/asjard/pkg/protobuf/metricspb"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MetricsAPI struct {
	reg prometheus.TransactionalGatherer
	metricspb.UnimplementedMetricsServer
}

func NewMetricsAPI(gather prometheus.Gatherer) *MetricsAPI {
	return &MetricsAPI{reg: prometheus.ToTransactionalGatherer(gather)}
}

// Fetch 获取系统指标
func (api *MetricsAPI) Fetch(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	rtx, ok := ctx.(*rest.Context)
	if ok {
		mfs, done, err := api.reg.Gather()
		defer done()
		if err != nil {
			logger.Error("gathering metrics fail", "err", err)
			return nil, ajerr.InternalServerError
		}
		contentType := expfmt.NegotiateIncludingOpenMetrics(rtx.ReadHeaderParams())
		rtx.Response.Header.Set(fasthttp.HeaderContentType, string(contentType))
		enc := expfmt.NewEncoder(rtx.RequestCtx.Response.BodyWriter(), contentType)
		for _, mf := range mfs {
			if err := enc.Encode(mf); err != nil {
				logger.Error("encoding and sending metric family fail", "err", err)
				return nil, ajerr.InternalServerError
			}
		}
		if closer, ok := enc.(expfmt.Closer); ok {
			if err := closer.Close(); err != nil {
				logger.Error("closer close fail", "err", err)
				return nil, ajerr.InternalServerError
			}
		}
		return nil, nil
	}

	return &emptypb.Empty{}, nil
}

func (MetricsAPI) RestServiceDesc() *rest.ServiceDesc {
	return &metricspb.MetricsRestServiceDesc
}
