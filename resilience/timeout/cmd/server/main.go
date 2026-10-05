// Command server is the flaky downstream: /slow sleeps ?ms=N
// (but quits early when the client goes away), /fast answers at
// once, /stats proves timeouts still cost server work (started vs
// finished diverge exactly by the cancelled count).
package main

import (
	"flag"
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

var started, finished atomic.Int64

// slow burns ?ms=N milliseconds of server time. Context-aware: a
// client timeout cancels the request context, and we stop before
// writing — the work already spent is still lost (see /stats).
func slow(w http.ResponseWriter, r *http.Request) {
	started.Add(1)

	ms := 1000

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

	finished.Add(1)

	_, _ = fmt.Fprintf(w, `{"slept_ms":%d}`, ms)
}

// fast answers immediately: the hedge target.
func fast(w http.ResponseWriter, _ *http.Request) {
	started.Add(1)
	finished.Add(1)

	_, _ = fmt.Fprint(w, `{"ok":true}`)
}

func stats(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, `{"started":%d,"finished":%d,"abandoned":%d}`,
		started.Load(), finished.Load(), started.Load()-finished.Load())
}

func main() {
	headerMs := flag.Int("header-ms", 5000, "max ms to read request headers (Slowloris guard)")

	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/slow", slow)
	mux.HandleFunc("/fast", fast)
	mux.HandleFunc("/stats", stats)

	addr := ":" + environ.Get("PORT", "8121")

	slog.Info("serving timeout lab", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(*headerMs) * time.Millisecond,
	}

	err := server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}
