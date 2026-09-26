// Command svc-a starts the trace: root span, W3C inject into the
// downstream call, svc-b continues it. One waterfall in Jaeger.
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

var tracer = otel.Tracer("svc-a")

func start(w http.ResponseWriter, r *http.Request) {
	// Root span: no incoming context on the edge (extract harmlessly anyway).
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	ctx, span := tracer.Start(ctx, "start")
	defer span.End()

	downstream := "http://" + environ.Get("SVC_B", "localhost:8111") + "/work"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, downstream, nil)

	// Inject: the traceparent header svc-b extracts. This is the manual
	// version of what otelhttp does automatically.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("downstream failed", "error", err)
		http.Error(w, "downstream unreachable", http.StatusBadGateway)

		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var downstreamBody map[string]string

	_ = json.NewDecoder(resp.Body).Decode(&downstreamBody)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"service":    "a",
		"trace_id":   span.SpanContext().TraceID().String(),
		"downstream": downstreamBody,
	})
}

func main() {
	ctx := context.Background()

	shutdown, err := tracing.Setup(ctx, "svc-a", environ.Get("OTEL_ENDPOINT", "localhost:4317"))
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/start", start)

	addr := ":" + environ.Get("PORT", "8110")

	slog.Info("serving otel svc-a", "addr", addr)

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
