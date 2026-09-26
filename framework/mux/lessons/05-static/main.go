// Lesson 5: static + edge cases. FileServer for assets, custom 404/405,
// trailing-slash behavior. Run: go run . (curl :8116/, :8116/nope,
// POST :8116/ (405), :8116/static/).
package main

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("home\n"))
	})
	// Static dir ./static (create it, drop an index.html to try).
	// StripPrefix so /static/a.css maps to ./static/a.css, not ./static/static/a.css.
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Custom 404: wrap mux, intercept miss before default NotFound writes.
	with404 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"nope"}` + "\n"))

			return
		}

		mux.ServeHTTP(w, r)
	})

	addr := ":8116"

	slog.Info("lesson 5: static", "addr", addr)

	_ = os.MkdirAll("./static", 0o750)

	server := &http.Server{
		Addr:              addr,
		Handler:           with404,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln("lesson failed:", err)
	}
}
