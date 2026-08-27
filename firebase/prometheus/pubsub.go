package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PubSubMetrics struct {
	messagePublishSuccess prometheus.Counter
	messagePublishFailure prometheus.Counter
}

func NewPubSubMetrics(reg prometheus.Registerer) *PubSubMetrics {
	hm := &PubSubMetrics{
		messagePublishSuccess: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "stripe_published_total",
			Help: "Stripe published total",
		}),
		messagePublishFailure: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "stripe_published_failed_total",
			Help: "Stripe published failed total",
		}),
	}
	return hm
}

func (em *PubSubMetrics) PublishSuccess() {
	em.messagePublishSuccess.Inc()
}

func (em *PubSubMetrics) PublishFailure() {
	em.messagePublishFailure.Inc()
}
