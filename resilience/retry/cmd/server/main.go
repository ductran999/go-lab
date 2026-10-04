// Command server fails the first ?fail=N hits on /flaky (then 200)
// and records every hit in 50ms buckets. /stats exposes the buckets:
// retries without jitter pile into the same buckets (tall spikes),
// jittered retries spread across them (low humps). /reset?fail=N
// rearms the budget and clears the buckets.
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

const bucketWidth = 50 * time.Millisecond

var (
	budget atomic.Int64
	hits   atomic.Int64

	mu      sync.Mutex
	start   = time.Now()
	buckets = map[int64]int{}
)

func bucket() {
	mu.Lock()

	defer mu.Unlock()

	buckets[int64(time.Since(start)/bucketWidth)]++
}

// flaky burns one hit: 500 while the failure budget lasts, 200 after.
func flaky(w http.ResponseWriter, _ *http.Request) {
	h := hits.Add(1)

	bucket()

	if h <= budget.Load() {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"ok":false}`)

		return
	}

	_, _ = fmt.Fprint(w, `{"ok":true}`)
}

// reset rearms: /reset?fail=N fails the next N hits from scratch.
func reset(w http.ResponseWriter, r *http.Request) {
	n := 0

	parsed, err := strconv.Atoi(r.URL.Query().Get("fail"))
	if err == nil && parsed >= 0 {
		n = parsed
	}

	budget.Store(int64(n))
	hits.Store(0)

	mu.Lock()

	start = time.Now()
	buckets = map[int64]int{}

	mu.Unlock()

	_, _ = fmt.Fprintf(w, `{"fail_next":%d}`, n)
}

func stats(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()

	defer mu.Unlock()

	out := map[string]any{
		"hits":    hits.Load(),
		"buckets": buckets,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		slog.Warn("stats encode failed", "error", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/flaky", flaky)
	mux.HandleFunc("/reset", reset)
	mux.HandleFunc("/stats", stats)

	addr := ":" + environ.Get("PORT", "8122")

	slog.Info("serving retry lab", "addr", addr)

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
