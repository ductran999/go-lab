// Command server is the expensive origin: /hot?ms=N burns N ms per
// query (ctx-aware) and counts every query it actually executes.
// /stats exposes the count — the stampede proof. /reset zeroes it.
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

var queries atomic.Int64

// hot simulates one expensive fetch: sleep, then answer. One
// execution per call — collapse the calls and this counter drops.
func hot(w http.ResponseWriter, r *http.Request) {
	queries.Add(1)

	ms := 300

	parsed, err := strconv.Atoi(r.URL.Query().Get("ms"))
	if err == nil && parsed >= 0 {
		ms = parsed
	}

	timer := time.NewTimer(time.Duration(ms) * time.Millisecond)

	defer timer.Stop()

	select {
	case <-r.Context().Done():
		return
	case <-timer.C:
	}

	_, _ = fmt.Fprintf(w, `{"item":"hot","slept_ms":%d}`, ms)
}

func stats(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, `{"queries":%d}`, queries.Load())
}

func reset(w http.ResponseWriter, _ *http.Request) {
	queries.Store(0)

	_, _ = fmt.Fprint(w, `{"reset":true}`)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hot", hot)
	mux.HandleFunc("/stats", stats)
	mux.HandleFunc("/reset", reset)

	addr := ":" + environ.Get("PORT", "8123")

	slog.Info("serving singleflight lab", "addr", addr)

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
