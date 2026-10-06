// Command server is the killable downstream: /flaky fails the first
// ?fail=N hits (then 200), /reset?fail=N rearms, /stats shows hits.
// Kill it (fail=forever) and the breaker opens; heal it (fail=0)
// and the half-open probe closes the circuit by itself.
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

var budget, hits atomic.Int64

// flaky burns one hit: 500 while the failure budget lasts, 200 after.
func flaky(w http.ResponseWriter, _ *http.Request) {
	h := hits.Add(1)

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

	_, _ = fmt.Fprintf(w, `{"fail_next":%d}`, n)
}

func stats(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, `{"hits":%d}`, hits.Load())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/flaky", flaky)
	mux.HandleFunc("/reset", reset)
	mux.HandleFunc("/stats", stats)

	addr := ":" + environ.Get("PORT", "8124")

	slog.Info("serving breaker lab", "addr", addr)

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
