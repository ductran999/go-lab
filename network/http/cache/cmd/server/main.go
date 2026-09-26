// Command server demonstrates HTTP caching primitives: validators
// (ETag + 304), freshness (max-age, origin untouched), and the
// hit counter proving which requests reached the origin at all.
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

var hits atomic.Int64

// version serves ?v=N with ETag "vN". Matching If-None-Match → 304
// (revalidation still reaches origin, but saves bytes).
func version(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query().Get("v")
	if v == "" {
		v = "1"
	}

	hits.Add(1)

	etag := `"` + v + `"`

	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)

		return
	}

	_, _ = fmt.Fprintf(w, `{"v":%q,"body":"%s"}`, v, string(make([]byte, 1024)))
}

// fresh serves a 60s-fresh response. Repeat requests within max-age
// never reach origin (counter stays flat) — freshness beats 304.
func fresh(w http.ResponseWriter, r *http.Request) {
	hits.Add(1)

	w.Header().Set("Cache-Control", "public, max-age=60")

	_, _ = fmt.Fprintf(w, `{"at":%q}`, time.Now().Format(time.RFC3339))
}

// nostore opts out entirely: every request hits origin, nothing kept.
func nostore(w http.ResponseWriter, r *http.Request) {
	hits.Add(1)

	w.Header().Set("Cache-Control", "no-store")

	_, _ = fmt.Fprintf(w, `{"at":%q}`, time.Now().Format(time.RFC3339))
}

func count(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, `{"hits":%d}`, hits.Load())
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/version", version)
	mux.HandleFunc("/fresh", fresh)
	mux.HandleFunc("/nostore", nostore)
	mux.HandleFunc("/hits", count)

	addr := ":" + environ.Get("PORT", "8099")

	slog.Info("serving cache lab", "addr", addr)

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
