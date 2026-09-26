// Package middleware holds shared HTTP middleware: logging.
// Chain as logging(recoverer(mux)) — stdlib has no Use(), wrapping is it.
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logging records method, path, status and duration per request.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &recorder{writer: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		slog.Info("http",
			"method", r.Method, "path", r.URL.Path,
			"status", rec.status, "dur", time.Since(start).Round(time.Millisecond))
	})
}

type recorder struct {
	writer http.ResponseWriter
	status int
}

func (r *recorder) Header() http.Header {
	return r.writer.Header()
}

func (r *recorder) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func (r *recorder) WriteHeader(status int) {
	r.status = status
	r.writer.WriteHeader(status)
}

// Flush forwards so streamed responses (SSE) keep working behind logging.
func (r *recorder) Flush() {
	if f, ok := r.writer.(http.Flusher); ok {
		f.Flush()
	}
}
