// Package server runs an http.Server with graceful shutdown, so mains
// stay wiring-only. Logging lives in middleware (single log line).
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

// Run serves handler until SIGINT/SIGTERM, then drains 10s.
// Returns the process exit code: main is just os.Exit(Run(...)).
// No logging wrapper here: services compose their own chain
// (one log line per request lives in middleware.TraceLog).
func Run(addr string, handler http.Handler) int {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()

		grace, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()

		shutdownErr := srv.Shutdown(grace)
		if shutdownErr != nil {
			slog.Error("drain failed", "error", shutdownErr)
		}
	}()

	slog.Info("serving", "addr", addr)

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "error", err)

		return 1
	}

	return 0
}
