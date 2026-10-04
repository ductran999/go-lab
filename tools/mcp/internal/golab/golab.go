// Package golab implements the go-lab MCP tools as plain functions
// (testable without the protocol): bench_run fires the OTel bench,
// metrics_query greps the Prometheus exposition, retry_stats dumps
// the retry lab buckets. Small outputs on purpose — every line
// lands in the agent's context.
package golab

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// maxLines caps tool output: agents pay per line, forever.
const maxLines = 50

var httpClient = &http.Client{Timeout: 10 * time.Second}

// BenchRun fires otel/bench.sh with N requests at concurrency C and
// returns its report (counter delta, buckets, one exemplar trace).
// Bounds keep a curious agent from benchmarking production to death.
func BenchRun(labDir string, n, c int) (string, error) {
	if n < 1 || n > 5000 || c < 1 || c > 200 {
		return "", fmt.Errorf("n in [1,5000], c in [1,200]: got n=%d c=%d", n, c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	defer cancel()

	cmd := exec.CommandContext(ctx, labDir+"/bench.sh", strconv.Itoa(n), strconv.Itoa(c))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("bench: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	return strings.TrimSpace(string(out)), nil
}

// MetricsQuery returns exposition lines containing pattern
// (e.g. "http_requests_total", "trace_id"). Empty pattern matches
// everything, so it is rejected — that dump is megabytes.
func MetricsQuery(metricsURL, pattern string) (string, error) {
	if strings.TrimSpace(pattern) == "" {
		return "", errors.New("pattern must not be empty")
	}

	body, err := getBody(metricsURL)
	if err != nil {
		return "", err
	}

	var hits []string

	for line := range strings.Lines(strings.TrimSpace(body)) {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, pattern) {
			hits = append(hits, line)
		}

		if len(hits) >= maxLines {
			hits = append(hits, fmt.Sprintf("... truncated at %d lines, narrow the pattern", maxLines))

			break
		}
	}

	if len(hits) == 0 {
		return fmt.Sprintf("no lines match %q", pattern), nil
	}

	return strings.Join(hits, "\n"), nil
}

// RetryStats returns the retry lab /stats document (hits + the
// per-50ms buckets that show lockstep vs spread).
func RetryStats(retryBase string) (string, error) {
	return getBody(strings.TrimSuffix(retryBase, "/") + "/stats")
}

func getBody(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("get %s: %w", url, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get %s: status %d", url, resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", url, err)
	}

	return string(raw), nil
}
