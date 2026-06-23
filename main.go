package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/rolldice", rolldice)
	mux.HandleFunc("/rolldice/{player}", rolldice)

	// Add HTTP instrumentation for the whole server
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

// Runs a HTTP server
//
// Returns an error if the server could not run or if the user called CTRL+C
func run() error {
	// Handle SIGINT (CTRL+C) gracefully
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	srv := &http.Server{
		Addr:         ":8000",
		BaseContext:  func(net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		log.Println("Running HTTP server on port 8000...")
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption
	select {
	case err := <-srvErr:
		// Error when starting HTTP server
		return err
	case <-ctx.Done():
		// Wait for first CTRL+C
		// Stop receiving signal notifications as soon as possible
		stop()
	}

	err = srv.Shutdown(context.Background())
	return err
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
