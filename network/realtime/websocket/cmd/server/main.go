// Command server demonstrates WebSocket origin validation: the
// handshake carries Origin, but no library checks it for you.
// Unknown origins get 403 before upgrade; known ones echo back.
// Allowlist comes from ORIGINS (comma-separated, default the demo
// page); PNA preflights get Allow-Private-Network when requested.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

// allowed reads ORIGINS once per handshake (cheap, always fresh).
func allowed(origin string) bool {
	for o := range strings.SplitSeq(environ.Get("ORIGINS", "http://localhost:8093"), ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}

	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Empty Origin (non-browser clients) rejected too.
	CheckOrigin: func(r *http.Request) bool {
		return allowed(r.Header.Get("Origin"))
	},
}

func echo(w http.ResponseWriter, r *http.Request) {
	// PNA preflight (Chrome: public page → local device): answer the
	// CORS preflight before any upgrade attempt.
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")

		if r.Header.Get("Access-Control-Request-Private-Network") == "true" {
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
		}

		w.WriteHeader(http.StatusNoContent)

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("upgrade rejected", "origin", r.Header.Get("Origin"), "error", err)

		return
	}

	defer func() {
		_ = conn.Close()
	}()

	for {
		kind, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// Demo clarity: answer ping with pong so sent and received
		// never look identical. Everything else echoes back untouched.
		if string(msg) == "ping" {
			msg = []byte("pong")
		}

		err = conn.WriteMessage(kind, msg)
		if err != nil {
			return
		}
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", echo)

	addr := ":" + environ.Get("PORT", "8094")

	slog.Info("serving websocket lab", "addr", addr)

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
