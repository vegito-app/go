package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type FirestoreMetrics struct {
	readSuccess        prometheus.Counter
	readFailure        prometheus.Counter
	writeSuccess       prometheus.Counter
	writeFailure       prometheus.Counter
	transactionFailure prometheus.Counter
}

func NewFirestoreMetrics(reg prometheus.Registerer) *FirestoreMetrics {
	hm := &FirestoreMetrics{
		readSuccess: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "firestore_read_total",
			Help: "Firestore read total",
		}),
		readFailure: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "firestore_read_failed_total",
			Help: "Firestore read failed total",
		}),
		writeSuccess: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "firestore_write_total",
			Help: "Firestore write total",
		}),
		writeFailure: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "firestore_write_failed_total",
			Help: "Firestore write failed total",
		}),
		transactionFailure: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "firestore_transaction_failed_total",
			Help: "Firestore transaction failed total",
		}),
	}
	return hm
}

func (em *FirestoreMetrics) ReadSuccess() {
	em.readSuccess.Inc()
}

func (em *FirestoreMetrics) ReadFailure() {
	em.readFailure.Inc()
}
func (em *FirestoreMetrics) WriteSuccess() {
	em.writeSuccess.Inc()
}

func (em *FirestoreMetrics) WriteFailure() {
	em.writeFailure.Inc()
}
func (em *FirestoreMetrics) TransactionFailure() {
	em.transactionFailure.Inc()
}
