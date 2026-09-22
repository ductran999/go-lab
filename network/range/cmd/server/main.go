// Command server demonstrates range requests: one readable text, full
// or partial. Stdlib ServeContent handles Range/If-Range/416 — the
// lab shows the wire shapes, not the parsing.
package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-lab/network/range/internal/blob"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func main() {
	// Readable text: numbered lines, any slice eyeball-verifiable.
	// FILE_MB sizes it (default 2, fast verify; crank up to watch
	// mid-download kills actually corrupt, e.g. FILE_MB=64).
	mb, err := strconv.Atoi(environ.Get("FILE_MB", "2"))
	if err != nil || mb < 1 {
		mb = 2
	}

	blob, digest := blob.Build(mb)

	mux := http.NewServeMux()
	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"lines-v1"`)
		w.Header().Set("Content-Digest", digest)
		http.ServeContent(w, r, "lines.txt", time.Now(), bytes.NewReader(blob))
	})

	mux.HandleFunc("/size", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"bytes":%d}`, len(blob))
	})

	addr := ":" + environ.Get("PORT", "8106")

	slog.Info("serving range lab", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}
