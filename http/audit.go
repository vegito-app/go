package http

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func auditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestTime := r.Context().Value(requestTimeKey).(int64)

		requestBody, _ := RequestBodyFromContext(r)
		// Logique d'audit ici, par exemple enregistrer l'URL et la méthode
		auditedFields := map[string]any{
			"method":      r.Method,
			"url":         r.URL.Path,
			"requestTime": requestTime,
			"requestBody": string(requestBody),
		}
		requestStartTime := time.Now()
		defer func() {
			auditedFields["duration"] = time.Since(requestStartTime)
			log.Info().Fields(auditedFields).Msg("Request received")
		}()
		// Appel du handler suivant
		next.ServeHTTP(w, r)
	})
}
