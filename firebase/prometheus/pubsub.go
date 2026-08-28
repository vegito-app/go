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
			Name: "pubsub_published_total",
			Help: "PubSub published total",
		}),
		messagePublishFailure: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "pubsub_published_failed_total",
			Help: "PubSub published failed total",
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
