package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"
	"github.com/spf13/viper"

	"github.com/rs/zerolog/log"
)

var config = viper.New()

const (
	portConfig = "port"
)

func init() {
	config.AutomaticEnv()
	config.SetDefault(portConfig, "8080")
}

func ListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Handle graceful shutdown
	waitShutdown := make(chan struct{})
	shutDownCtx, cancelWaitShutdown := context.WithCancel(context.Background())
	defer func() {
		cancelWaitShutdown()
		<-waitShutdown
	}()
	var serverError error
	// Start a goroutine that will shutdown the HTTP server when the context is done or
	// when an OS interruption signal is received.
	go func() {
		defer close(waitShutdown)

		// Create a channel to receive OS signals
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		// Wait for either the context to be done or an OS signal to be received
		select {
		case <-ctx.Done():
			log.Info().Msg("Context done, shutting down HTTP server")
			// Shutdown the server
			shutdown(ctx, server)

		case <-c:
			log.Info().Msg("Received OS interruption signal, shutting down HTTP server")
			// shutdown the server
			shutdown(ctx, server)

		case <-shutDownCtx.Done():
			log.Error().Err(serverError).Msg("HTTP server has been stopped internally")
		}
	}()

	// This goroutine will shutdown the HTTP server when the context is done or when an
	// OS interruption signal is received. It will wait for either the context to be
	// done or an OS signal to be received, then shutdown the server. If the server
	// has already been stopped internally, it will log an error message. If the shutdown
	// is successful, it will log an info message.

	serverError = server.ListenAndServe()
	return serverError
}

func shutdown(ctx context.Context, server *http.Server) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Msg("HTTP server shutdown, error: " + err.Error())
		return
	}
	log.Info().Msg("HTTP server has gracefully shutdown")
}

// StartAPI creates a new instance of apiv1.Service.
func StartAPI(ctx context.Context, mux http.Handler, metricsProbe Metrics) error {
	handler := ApplyMiddleware(
		mux,
		MetricsMiddleware(metricsProbe),
		RequestBodyMiddleware,
		AuditMiddleware,
		CorsMiddleware,
	)
	port := config.GetString(portConfig)
	if err := ListenAndServe(ctx, "0.0.0.0:"+port, handler); err != nil {
		return fmt.Errorf("HTTP listenAndServe: %w", err)
	}
	fmt.Print("See you the next time ! Bye")
	return nil
}
