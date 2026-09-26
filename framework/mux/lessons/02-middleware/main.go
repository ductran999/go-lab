// Lesson 2: middleware. Stdlib has no Use(): chain by wrapping.
// Order reads outside-in: logging(recover(auth(mux))) runs logging
// first, auth last, handler innermost. Per-route: wrap one handler.
// Run: go run . (curl :8113/open and :8113/secret).
package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	hdl := &Handler{}

	mux.HandleFunc("GET /public/login", hdl.Login)

	// Per-route wrap: auth applies to /me only. Handle (not
	// HandleFunc) takes a Handler, so wrapped chains fit directly.
	mux.Handle("GET /me", auth(http.HandlerFunc(hdl.Me)))

	addr := ":8113"

	slog.Info("lesson 2: middleware", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           logging(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln("lesson failed:", err)
	}
}

type Handler struct{}

func (h *Handler) Login(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("login page\n"))
}

func (h *Handler) Me(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("hello daniel\n"))
}

// logging is global: every request in and out with duration.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)
		slog.Info("http", "method", r.Method, "path", r.URL.Path, "dur", time.Since(start))
	})
}

// auth is per-route: only /secret checks the token.
func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer demo" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	})
}
