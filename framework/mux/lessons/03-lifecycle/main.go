// Lesson 3: lifecycle. Timeouts bound every phase (Slowloris dies
// here); Shutdown drains in-flight requests on SIGINT/SIGTERM.
// Run: go run . (curl :8114/slow, then Ctrl-C mid-request to watch drain).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /fast", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fast\n"))
	})
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("slow handler started", "remote", r.RemoteAddr)
		// Respect client disconnect: work stops when they go away.
		select {
		case <-r.Context().Done():
			slog.Warn("user left before work")
			return
		case <-time.After(10 * time.Second):
			_, _ = w.Write([]byte("slow done\n"))
		}
	})

	server := &http.Server{
		Addr:              ":8114",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,  // headers must arrive fast
		ReadTimeout:       10 * time.Second, // whole request read budget
		WriteTimeout:      15 * time.Second, // whole response write budget
		IdleTimeout:       60 * time.Second, // keep-alive idle cap
	}

	// Shutdown: stop accepting, let in-flight finish (30s grace).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// main blocks here until the drain actually finishes: without this,
	// ListenAndServe returns the instant Shutdown is *called* (not when
	// it completes), main falls through, and the process exits mid-drain.
	shutdownDone := make(chan struct{})

	go func() {
		defer close(shutdownDone)

		<-ctx.Done()
		slog.Info("draining...")

		// Fresh deadline, inherited (not tied to the signal itself).
		grace, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		shutdownErr := server.Shutdown(grace)
		if shutdownErr != nil {
			slog.Error("drain failed", "error", shutdownErr)
		}
	}()

	slog.Info("lesson 3: lifecycle", "addr", server.Addr)

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("lesson failed", "error", err)
		return
	}

	<-shutdownDone
}
