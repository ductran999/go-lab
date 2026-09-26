// Command server demonstrates SSE: a long-lived HTTP response the
// server writes events into. Client sends nothing after subscribe;
// EventSource auto-reconnects and resumes via Last-Event-ID.
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

// events streams an incrementing counter forever. ?kill_after=N closes
// the connection after N events so EventSource reconnects and resumes
// from Last-Event-ID. ?tick_ms slows the ticker for demos.
func events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8095")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)

		return
	}

	start := 1

	if last := r.Header.Get("Last-Event-ID"); last != "" {
		n, convErr := strconv.Atoi(last)
		if convErr == nil {
			start = n + 1
		}
	}

	killAfter := 0

	if q := r.URL.Query().Get("kill_after"); q != "" {
		n, convErr := strconv.Atoi(q)
		if convErr == nil {
			killAfter = n
		}
	}

	tickMS := 1000

	if q := r.URL.Query().Get("tick_ms"); q != "" {
		n, convErr := strconv.Atoi(q)
		if convErr == nil && n >= 100 {
			tickMS = n
		}
	}

	ticker := time.NewTicker(time.Duration(tickMS) * time.Millisecond)
	defer ticker.Stop()

	sent := 0

	for id := start; ; id++ {
		select {
		case <-r.Context().Done():
			return
		case now := <-ticker.C:
			_, _ = fmt.Fprintf(w, "id: %d\nevent: tick\ndata: {\"n\":%d,\"at\":%q}\n\n", id, id, now.Format(time.RFC3339))

			flusher.Flush()

			sent++

			if killAfter > 0 && sent >= killAfter {
				return
			}
		}
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", events)

	addr := ":" + environ.Get("PORT", "8096")

	slog.Info("serving sse lab", "addr", addr)

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
