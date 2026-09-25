// Command server speaks HTTP/3 (QUIC + TLS): one unary endpoint and
// one finite event stream. Certs are shared with the grpc lab demo CA
// (localhost identity) — regenerate there when expired.
package main

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/quic-go/quic-go/http3"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func hello(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, `{"proto":%q,"msg":"hello over QUIC"}`+"\n", r.Proto)
}

func stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")

	for i := 1; i <= 5; i++ {
		select {
		case <-r.Context().Done():
			return
		default:
			_, _ = fmt.Fprintf(w, "id: %d\ndata: {\"n\":%d}\n\n", i, i)

			flusher.Flush()
		}
	}

	_, _ = fmt.Fprint(w, "event: done\ndata: {}\n\n")

	flusher.Flush()
}

func main() {
	cert, err := tls.LoadX509KeyPair(
		environ.Get("TLS_CERT", "../grpc/certs/server.crt"),
		environ.Get("TLS_KEY", "../grpc/certs/server.key"),
	)
	if err != nil {
		fail(fmt.Errorf("tls: %w", err))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/stream", stream)

	addr := ":" + environ.Get("PORT", "8108")

	server := &http3.Server{
		Handler:   mux,
		Addr:      addr,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
	}

	slog.Info("serving http3 lab", "addr", addr, "note", "QUIC requires TLS; certs shared with grpc lab")

	serveErr := server.ListenAndServe()
	if serveErr != nil {
		fail(serveErr)
	}
}
