// Command server exposes go-lab to agents over MCP stdio: bench_run
// fires the OTel bench, metrics_query greps the exposition,
// retry_stats dumps the retry buckets. Run from the repo root (or
// set OTEL_LAB_DIR); logs go to stderr, stdout is pure protocol.
package main

import (
	"log/slog"
	"os"

	"go-lab/tools/mcp/internal/golab"
	"go-lab/tools/mcp/internal/mcp"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	labDir := getenv("OTEL_LAB_DIR", "observability/otel")
	metricsURL := getenv("METRICS_URL", "http://localhost:2112/metrics")
	retryBase := getenv("RETRY_BASE", "http://localhost:8122")

	srv := mcp.New("go-lab", "0.1.0")

	srv.Add(mcp.Tool{
		Name:        "bench_run",
		Description: "Fire N requests at the OTel lab and report delta, buckets, one exemplar trace.",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"n": map[string]any{"type": "integer", "minimum": 1, "maximum": 5000, "default": 200},
				"c": map[string]any{"type": "integer", "minimum": 1, "maximum": 200, "default": 20},
			},
		},
	}, func(args map[string]any) (string, error) {
		return golab.BenchRun(labDir, mcp.IntArg(args, "n", 200, 1, 5000), mcp.IntArg(args, "c", 20, 1, 200))
	})

	srv.Add(mcp.Tool{
		Name:        "metrics_query",
		Description: "Grep the Prometheus exposition for lines containing pattern (metric names, trace_id, buckets).",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"pattern": map[string]any{"type": "string"}},
			"required":   []string{"pattern"},
		},
	}, func(args map[string]any) (string, error) {
		return golab.MetricsQuery(metricsURL, mcp.StrArg(args, "pattern", ""))
	})

	srv.Add(mcp.Tool{
		Name:        "retry_stats",
		Description: "Dump the retry lab /stats: total hits plus per-50ms buckets (lockstep spikes vs jittered humps).",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(_ map[string]any) (string, error) {
		return golab.RetryStats(retryBase)
	})

	slog.Info("mcp go-lab serving on stdio")

	srv.Serve(os.Stdin, os.Stdout)
}
