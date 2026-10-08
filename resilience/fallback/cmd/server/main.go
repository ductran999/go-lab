// Command server is the killable primary: /flaky fails the first
// ?fail=N hits (then 200), /reset?fail=N rearms, /stats counts hits.
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

var budget, hits, dieAfter atomic.Int64

// flaky burns one hit: 500 while the failure budget lasts, 200 after —
// unless dieafter is set, past which the primary dies mid-run (the
// honest stale story: seed first, fail later, one process).
func flaky(w http.ResponseWriter, _ *http.Request) {
	h := hits.Add(1)

	if h <= budget.Load() {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"ok":false}`)

		return
	}

	if d := dieAfter.Load(); d >= 0 && h > d {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, `{"ok":false,"why":"died mid-run"}`)

		return
	}

	_, _ = fmt.Fprintf(w, `{"price":%d}`, 100+h)
}

// reset rearms: /reset?fail=N fails the next N hits from scratch,
// ?dieafter=M succeeds M hits then dies (default -1: never).
func reset(w http.ResponseWriter, r *http.Request) {
	n := 0

	parsed, err := strconv.Atoi(r.URL.Query().Get("fail"))
	if err == nil && parsed >= 0 {
		n = parsed
	}

	m := -1

	die, err := strconv.Atoi(r.URL.Query().Get("dieafter"))
	if err == nil && die >= 0 {
		m = die
	}

	budget.Store(int64(n))
	dieAfter.Store(int64(m))
	hits.Store(0)

	_, _ = fmt.Fprintf(w, `{"fail_next":%d,"die_after":%d}`, n, m)
}

func stats(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, `{"hits":%d}`, hits.Load())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/flaky", flaky)
	mux.HandleFunc("/reset", reset)
	mux.HandleFunc("/stats", stats)

	addr := ":" + environ.Get("PORT", "8127")

	slog.Info("serving fallback lab", "addr", addr)

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
