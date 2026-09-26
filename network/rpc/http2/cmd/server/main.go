// Command server speaks REST JSON over H2C (cleartext HTTP/2):
// same dialect as any H1 REST lab, faster vehicle. No TLS here —
// curl needs --http2-prior-knowledge for h2c (or use cmd/client).
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

type todo struct {
	ID   int64  `json:"id"`
	Task string `json:"task"`
	Done bool   `json:"done"`
}

type store struct {
	mu     sync.Mutex
	todos  map[int64]todo
	nextID int64
}

func (s *store) list(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]todo, 0, len(s.todos))
	for _, t := range s.todos {
		out = append(out, t)
	}

	writeJSON(w, http.StatusOK, out)
}

func (s *store) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Task string `json:"task"`
	}

	decodeErr := json.NewDecoder(r.Body).Decode(&body)
	if decodeErr != nil || body.Task == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "task is required"})

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++

	t := todo{ID: s.nextID, Task: body.Task}
	s.todos[t.ID] = t

	writeJSON(w, http.StatusCreated, t)
}

func (s *store) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/todos/"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad id"})

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.todos[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})

		return
	}

	writeJSON(w, http.StatusOK, t)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	st := &store{todos: make(map[int64]todo)}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", st.list)
	mux.HandleFunc("POST /todos", st.create)
	mux.HandleFunc("GET /todos/{id}", st.get)

	addr := ":" + environ.Get("PORT", "8109")

	slog.Info("serving http2-rest lab", "addr", addr, "note", "h2c cleartext: same REST, H2 vehicle")

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	cert, key := environ.Get("TLS_CERT", ""), environ.Get("TLS_KEY", "")
	if cert != "" && key != "" {
		// HTTPS path: ALPN negotiates H2 automatically (no Protocols
		// tuning needed — TLS implies encrypted H2). Browser-compatible.
		slog.Info("serving https+h2", "addr", addr)

		err := server.ListenAndServeTLS(cert, key)
		if err != nil {
			fail(err)
		}

		return
	}

	// H2C path (local demo): unencrypted HTTP/2 via stdlib Protocols.
	server.Protocols = &http.Protocols{}
	server.Protocols.SetHTTP1(true)
	server.Protocols.SetUnencryptedHTTP2(true)

	slog.Info("serving h2c", "addr", addr)

	err := server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}
