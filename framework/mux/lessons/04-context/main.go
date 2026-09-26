// Lesson 4: context + values. Request-scoped data rides r.Context():
// middleware sets, handlers read. Typed keys only (never bare strings).
// Timeout/cancel propagate: downstream work shares the request deadline.
// Run: go run . (curl :8115/whoami and :8115/slow-ctx).
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"
)

// ctxKey is unexported: no package can collide or forge our values.
type ctxKey string

const reqIDKey ctxKey = "reqID"

// withReqID stamps each request with an id (middleware).
func withReqID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), reqIDKey, id)))
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /whoami", func(w http.ResponseWriter, r *http.Request) {
		// Same request, same context: value set upstream is visible here.
		id, _ := r.Context().Value(reqIDKey).(string)
		_, _ = fmt.Fprintf(w, "reqID=%s\n", id)
	})
	mux.HandleFunc("GET /slow-ctx", func(w http.ResponseWriter, r *http.Request) {
		// Child deadline: at most 3s even if the client waits longer.
		// Cancel propagates both ways (client gone OR timeout).
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		select {
		case <-ctx.Done():
			slog.Info("slow-ctx stopped", "reason", ctx.Err(), "reqID", r.Context().Value(reqIDKey))
			return
		case <-time.After(10 * time.Second):
			_, _ = w.Write([]byte("slow done (never here: deadline is 3s)\n"))
		}
	})

	addr := ":8115"

	slog.Info("lesson 4: context", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           withReqID(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalln("lesson failed:", err)
	}
}
