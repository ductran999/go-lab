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

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	"go-lab/observability/otel/internal/metrics"
	"go-lab/observability/otel/internal/middleware"
	"go-lab/observability/otel/internal/requestid"
	"go-lab/observability/otel/internal/server"
	"go-lab/observability/otel/internal/tracing"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func work(w http.ResponseWriter, r *http.Request) {
	// Span already started by otelhttp (child of svc-a's): read it.
	// Baggage rode along with traceparent: tenant/user without re-auth.
	span := trace.SpanFromContext(r.Context())
	bag := baggage.FromContext(r.Context())

	slog.Info("work",
		"trace.id", span.SpanContext().TraceID().String(),
		"tenant.id", bag.Member("tenant.id").Value(),
		"user.id", bag.Member("user.id").Value())

	// Pretend something measurable happens here.
	select {
	case <-r.Context().Done():
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

	shutdown, err := tracing.Setup(ctx,
		tracing.NewServiceInfo("svc-b", "1.0.0", "pipeline"),
		environ.Get("OTEL_ENDPOINT", "localhost:4317"),
	)
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	prom, err := metrics.Setup()
	if err != nil {
		fail(err)
	}

	// Metrics on :2113 (svc-a holds :2112): same OTel-wired handler.
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", prom)

		maddr := ":" + environ.Get("METRICS_PORT", "2113")

		slog.Info("serving metrics", "addr", maddr)

		msrv := &http.Server{
			Addr:              maddr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}

		_ = msrv.ListenAndServe()
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/work", work)

	protected := http.NewServeMux()
	protected.Handle("/work", middleware.Auth(middleware.Logging(mux)))

	server.Run(
		":"+environ.Get("PORT", "8111"),
		requestid.Ensure(otelhttp.NewHandler(protected, "svc-b")),
	)
}
