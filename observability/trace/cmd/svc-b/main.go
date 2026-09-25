// Command svc-b is the downstream service: it reads traceparent,
// logs under the shared trace ID, and reports it back.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go-lab/observability/trace/internal/trace"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func work(w http.ResponseWriter, r *http.Request) {
	ctx := trace.Extract(r)

	slog.Info("svc-b work", "trace_id", ctx.TraceID, "span_id", ctx.ParentID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"service": "b", "trace_id": ctx.TraceID})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/work", work)

	addr := ":" + environ.Get("PORT", "8101")

	slog.Info("serving svc-b", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}
