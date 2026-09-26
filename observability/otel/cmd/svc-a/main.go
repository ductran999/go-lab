// Command svc-a starts the trace: root span, W3C inject into the
// downstream call, svc-b continues it. One waterfall in Jaeger.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

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

// token mints a demo JWT (tenant_id + user_id, 1h). Demo issuance:
// no login, no rotation — the transport lesson matters, not the PKI.
func token(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TenantID string `json:"tenant_id"`
		UserID   string `json:"user_id"`
	}

	_ = json.NewDecoder(r.Body).Decode(&body)

	if body.TenantID == "" {
		body.TenantID = "demo-tenant"
	}

	if body.UserID == "" {
		body.UserID = "demo-user"
	}

	signed, err := middleware.Mint(body.TenantID, body.UserID)
	if err != nil {
		http.Error(w, "mint failed", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": signed})
}

// start serves the traced endpoint. Span already started by otelhttp:
// read it, never start one here.
func start(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	downstream := "http://" + environ.Get("SVC_B", "localhost:8111") + "/work"

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, downstream, nil)

	// Forward identity downstream untouched: Bearer (svc-b verifies)
	// and the request id (svc-b logs the same join key).
	if auth := r.Header.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	if id := requestid.Of(ctx); id != "" {
		req.Header.Set("X-Request-Id", id)
	}

	// otelhttp transport injects traceparent automatically.
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	resp, err := client.Do(req)
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
		"success": true,
		"data": map[string]any{
			"service":    "a",
			"downstream": downstreamBody,
		},
		"trace_id":   span.SpanContext().TraceID().String(),
		"request_id": requestid.Of(ctx),
	})
}

func main() {
	ctx := context.Background()

	shutdown, err := tracing.Setup(ctx,
		tracing.NewServiceInfo("svc-a", "1.0.0", "pipeline"),
		environ.Get("OTEL_ENDPOINT", "localhost:4317"),
	)
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/start", start)

	// /token stays outside Auth: you can't require a token to mint one.
	// Chain: otelhttp (span) → Auth (identity) → Logging (the one
	// log line) → mux.
	public := http.NewServeMux()
	public.Handle("/start", middleware.Auth(middleware.Logging(mux)))
	public.HandleFunc("/token", token)

	server.Run(":"+environ.Get("PORT", "8110"),
		requestid.Ensure(otelhttp.NewHandler(public, "svc-a")))
}
