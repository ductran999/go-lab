// Command server is the plain downstream: /work?ms=N burns N ms
// and counts hits. No bucket here — throttling lives in the client
// (this pillar is client-side); the server just does honest work.
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

var hits atomic.Int64

// work burns ?ms=N unless the client goes away.
func work(w http.ResponseWriter, r *http.Request) {
	hits.Add(1)

	ms := 50

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

	_, _ = fmt.Fprintf(w, `{"slept_ms":%d}`, ms)
}

func stats(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, `{"hits":%d}`, hits.Load())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/work", work)
	mux.HandleFunc("/stats", stats)

	addr := ":" + environ.Get("PORT", "8126")

	slog.Info("serving ratelimit lab", "addr", addr)

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
