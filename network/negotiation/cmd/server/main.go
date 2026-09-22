// Command server demonstrates content negotiation: one endpoint,
// three representations. Accept picks; unsupported asks get 406.
// Quality values (q=) rank preferences when several match.
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

type todo struct {
	ID   int
	Task string
	Done bool
}

var todos = []todo{{ID: 1, Task: "write lab", Done: true}, {ID: 2, Task: "verify live", Done: false}}

// best picks the first supported type by client q-ranking. Empty
// means nothing matched.
func best(accept string) string {
	supported := []string{"application/json", "text/csv", "text/plain"}

	bestQ := -1.0
	bestName := ""

	for part := range strings.SplitSeq(accept, ",") {
		media := strings.TrimSpace(strings.Split(part, ";")[0])
		if media == "" || media == "*/*" {
			continue
		}

		q := 1.0

		if i := strings.Index(part, "q="); i >= 0 {
			_, _ = fmt.Sscanf(strings.TrimSpace(part[i+2:]), "%f", &q)
		}

		for _, s := range supported {
			if media == s && q > bestQ {
				bestQ = q
				bestName = s
			}
		}
	}

	return bestName
}

func data(w http.ResponseWriter, r *http.Request) {
	// Vary: caches must key on Accept, or a CSV poisons JSON clients.
	w.Header().Set("Vary", "Accept")

	accept := r.Header.Get("Accept")

	// Absent Accept = client takes anything: serve the default.
	// Present but unmatchable = 406 with the supported list.
	format := "application/json"
	if accept != "" {
		format = best(accept)
	}

	switch format {
	case "text/csv":
		w.Header().Set("Content-Type", "text/csv")
		_, _ = fmt.Fprintln(w, "id,task,done")

		for _, t := range todos {
			_, _ = fmt.Fprintf(w, "%d,%s,%t\n", t.ID, t.Task, t.Done)
		}
	case "text/plain":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		for _, t := range todos {
			_, _ = fmt.Fprintf(w, "[%d] %s done=%t\n", t.ID, t.Task, t.Done)
		}
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"id":1,"task":"write lab","done":true},{"id":2,"task":"verify live","done":false}]`)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = fmt.Fprintf(w, `{"error":"supported: application/json, text/csv, text/plain; got %q"}`, accept)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/data", data)

	addr := ":" + environ.Get("PORT", "8102")

	slog.Info("serving negotiation lab", "addr", addr)

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
