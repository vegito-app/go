package http

import (
	"context"
	"net/http"
	"time"
)

// Metrics defines the interface for HTTP metrics
type Metrics interface {

	// RequestsTotalInc increments the total number of requests
	RequestsTotalInc()
	// ObserveRequestDuration records the duration of a request
	ObserveRequestDuration(float64)
	// RequestsInFlightInc increments the number of requests in flight
	RequestsInFlightInc()
	// RequestsInFlightDec decrements the number of requests in flight
	RequestsInFlightDec()
}

type requestTimeContextKey string

const requestTimeKey requestTimeContextKey = "requestTime"

func MetricsMiddleware(metricsProbe Metrics) Middleware {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			requestTime := r.Context().Value(requestTimeKey)
			if requestTime == nil {
				// Si le contexte ne contient pas de valeur pour "requestTime", on peut définir une valeur par défaut
				requestTime = r.Context().Value("requestUnixTime")
			}
			// On peut aussi vérifier si requestTime est de type time.Time ou int64 selon le contexte
			if t, ok := requestTime.(time.Time); ok {
				// Si c'est un time.Time, on peut l'utiliser directement
				r = r.WithContext(context.WithValue(r.Context(), requestTimeKey, t.Unix()))
			} else if t, ok := requestTime.(int64); ok {
				// Si c'est un int64, on peut l'utiliser directement
				r = r.WithContext(context.WithValue(r.Context(), requestTimeKey, t))
			} else {
				// Si ce n'est ni l'un ni l'autre, on peut définir une valeur par défaut
				r = r.WithContext(context.WithValue(r.Context(), requestTimeKey, time.Now().Unix()))
			}
			metricsProbe.RequestsTotalInc()

			metricsProbe.RequestsInFlightInc()
			defer metricsProbe.RequestsInFlightDec()

			requestStartTime := time.Now()
			defer func() {
				metricsProbe.ObserveRequestDuration(time.Since(requestStartTime).Seconds())
			}()
			// Appel du handler suivant
			next.ServeHTTP(w, r)
		})
	}
}
