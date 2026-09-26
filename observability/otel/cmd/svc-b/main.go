// Command svc-b is the downstream span: it extracts the W3C context
// the caller injected, works in a child span, and returns.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"go-lab/observability/otel/internal/middleware"
	"go-lab/observability/otel/internal/tracing"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

var tracer = otel.Tracer("svc-b")

func work(w http.ResponseWriter, r *http.Request) {
	// Extract: continues the trace the caller started (same trace ID).
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	ctx, span := tracer.Start(ctx, "work")
	defer span.End()

	// Pretend something measurable happens here.
	select {
	case <-ctx.Done():
	case <-time.After(50 * time.Millisecond):
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"service":  "b",
		"trace_id": span.SpanContext().TraceID().String(),
	})
}

func main() {
	ctx := context.Background()

	shutdown, err := tracing.Setup(ctx, "svc-b", environ.Get("OTEL_ENDPOINT", "localhost:4317"))
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/work", work)

	addr := ":" + environ.Get("PORT", "8111")

	slog.Info("serving otel svc-b", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           middleware.Logging(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}
