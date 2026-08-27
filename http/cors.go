package http

import (
	"net/http"
)

// CORSMiddleware définit un middleware pour gérer le CORS
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Autoriser les requêtes depuis n'importe quelle origine (en dev)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Autoriser les méthodes HTTP spécifiques
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Autoriser les en-têtes spécifiques
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Gérer la pré-vérification des requêtes OPTIONS (préflight requests)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Passer à l'étape suivante
		next.ServeHTTP(w, r)
	})
}