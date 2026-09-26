// Lesson 1: routing. Go 1.22+ patterns: [METHOD ][HOST]/path,
// {name} captures one segment, {$} anchors the end.
// Precedence: most specific wins (static > wildcard, method match first).
// Run: go run . (then curl the printed URLs).
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	hdl := &Handler{}

	mux.HandleFunc("GET /todos", hdl.ListTodos)
	mux.HandleFunc("GET /todos/{id}", hdl.GetTodo)
	// {$} anchors: /files/ matches, /files/a/b does not.
	// /files (no slash) 301-redirects to /files/: ServeMux cleans the
	// path, finds the registered /files/ subtree, and canonicalizes.
	mux.HandleFunc("/files/{$}", hdl.FilesRoot)

	addr := ":8112"

	slog.Info("lesson 1: routing", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln("start server error", err)
	}
}

var todos = []Todo{
	{ID: "1", Description: "learn mux"},
	{ID: "2", Description: "write todo app"},
}

type Handler struct{}

type Todo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

func (h *Handler) ListTodos(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, todos)
}

// FilesRoot serves the anchored subtree root as plain text.
func (h *Handler) FilesRoot(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprint(w, "files root\n")
	if err != nil {
		slog.Warn("client gone mid-write", "error", err)
	}
}

func (h *Handler) GetTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	for _, td := range todos {
		if td.ID == id {
			writeJSON(w, http.StatusOK, td)

			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{"message": "not found"})
}

// writeJSON is the single exit for JSON responses: marshal first (safe
// to 500: nothing written yet), then headers, status, body. Write
// failures mean the client is gone — log and move on. Server side
// never Closes w: return and the server reaps it.
func writeJSON(w http.ResponseWriter, status int, v any) {
	respBody, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "encode failed", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	_, err = w.Write(respBody)
	if err != nil {
		slog.Warn("client gone mid-write", "error", err)
	}
}
