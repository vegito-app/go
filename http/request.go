package http

import (
	"context"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

type RequestMetrics struct {
	requestsTotal    prometheus.Counter
	requestsInFlight prometheus.Gauge
	requestDuration  prometheus.Histogram
}

func NewRequestMetrics(reg prometheus.Registerer) *RequestMetrics {
	hm := &RequestMetrics{
		requestsTotal: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		}),
		requestsInFlight: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed.",
		}),
		requestDuration: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}),
	}
	return hm
}

func (m *RequestMetrics) RequestsTotalInc() {
	m.requestsTotal.Inc()
}

func (m *RequestMetrics) ObserveRequestDuration(d float64) {
	m.requestDuration.Observe(d)
}

func (m *RequestMetrics) RequestsInFlightInc() {
	m.requestsInFlight.Inc()
}

func (m *RequestMetrics) RequestsInFlightDec() {
	m.requestsInFlight.Dec()
}

const requestBodyKey = "requestBody"

func setRequestBodyInContext(r *http.Request, body []byte) *http.Request {
	ctx := r.Context()
	return r.WithContext(context.WithValue(ctx, requestBodyKey, body))
}

func RequestBodyFromContext(r *http.Request) ([]byte, bool) {
	ctx := r.Context()
	body, ok := ctx.Value(requestBodyKey).([]byte)
	if !ok {
		return nil, false
	}
	return body, true
}

func RequestBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read request body")
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		r = setRequestBodyInContext(r, requestBody)

		next.ServeHTTP(w, r)
	})
}
