// Lesson 1: routing. Go 1.22+ patterns: [METHOD ][HOST]/path,
// {name} captures one segment, {$} anchors the end.
// Precedence: most specific wins (static > wildcard, method match first).
// Run: go run . (then curl the printed URLs).
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func fail(err error) {
	slog.Error("lesson failed", "error", err)

	os.Exit(1)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /todos", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "list\n")
	})
	// Static beats wildcard: /todos/new never reaches {id}.
	mux.HandleFunc("GET /todos/new", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "new form\n")
	})
	mux.HandleFunc("GET /todos/{id}", func(rw http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(rw, "todo id=%s\n", r.PathValue("id"))
	})
	// {$} anchors: /files/ matches, /files/a/b does not.
	mux.HandleFunc("/files/{$}", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "files root\n")
	})

	addr := ":8112"

	slog.Info("lesson 1: routing", "addr", addr)

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
