// Command svc-a is the entry service: it extracts (or starts) the
// trace, calls svc-b with a child context, and returns both halves.
// One trace ID across two logs and one response.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go-lab/network/trace/internal/trace"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func start(w http.ResponseWriter, r *http.Request) {
	ctx := trace.Extract(r)

	slog.Info("svc-a start", "trace_id", ctx.TraceID, "span_id", ctx.ParentID)

	downstream := "http://" + environ.Get("SVC_B", "localhost:8101") + "/work"

	req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, downstream, nil)
	req.Header.Set(trace.Header, ctx.Child().Value())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("svc-a downstream failed", "trace_id", ctx.TraceID, "error", err)
		http.Error(w, "downstream unreachable", http.StatusBadGateway)

		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var downstreamBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&downstreamBody)

	slog.Info("svc-a done", "trace_id", ctx.TraceID, "downstream_trace", downstreamBody["trace_id"])

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"service":    "a",
		"trace_id":   ctx.TraceID,
		"downstream": downstreamBody,
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", start)

	addr := ":" + environ.Get("PORT", "8100")

	slog.Info("serving svc-a", "addr", addr)

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
