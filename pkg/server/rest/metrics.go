package rest

import (
	"context"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsAPI struct{}

func (MetricsAPI) Fetch(ctx context.Context) {
	promhttp.Handler()
}
