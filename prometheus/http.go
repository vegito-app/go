package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsService struct {
	metrics Metrics
}

type Metrics interface {
	prometheus.Gatherer
}

func NewMetricsService(mux *http.ServeMux, metricsProvider Metrics) (*MetricsService, error) {
	service := &MetricsService{
		metrics: metricsProvider,
	}

	mux.Handle("GET /metrics", promhttp.HandlerFor(metricsProvider, promhttp.HandlerOpts{}))

	return service, nil
}
