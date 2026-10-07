// Command server hosts two downstreams: /slow?ms=N burns N ms,
// /fast answers at once. /stats counts per-endpoint hits.
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

var slowHits, fastHits atomic.Int64

// sleep burns ?ms=N unless the client goes away.
func sleep(w http.ResponseWriter, r *http.Request, ms int) {
	timer := time.NewTimer(time.Duration(ms) * time.Millisecond)

	defer timer.Stop()

	select {
	case <-r.Context().Done():
		return
	case <-timer.C:
	}

	_, _ = fmt.Fprintf(w, `{"slept_ms":%d}`, ms)
}

func slow(w http.ResponseWriter, r *http.Request) {
	slowHits.Add(1)

	ms := 1000

	parsed, err := strconv.Atoi(r.URL.Query().Get("ms"))
	if err == nil && parsed >= 0 {
		ms = parsed
	}

	sleep(w, r, ms)
}

func fast(w http.ResponseWriter, _ *http.Request) {
	fastHits.Add(1)

	_, _ = fmt.Fprint(w, `{"ok":true}`)
}

func stats(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, `{"slow":%d,"fast":%d}`, slowHits.Load(), fastHits.Load())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", slow)
	mux.HandleFunc("/fast", fast)
	mux.HandleFunc("/stats", stats)

	addr := ":" + environ.Get("PORT", "8125")

	slog.Info("serving bulkhead lab", "addr", addr)

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
