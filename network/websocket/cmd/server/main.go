// Command server demonstrates WebSocket origin validation: the
// handshake carries Origin, but no library checks it for you.
// Unknown origins get 403 before upgrade; known ones echo back.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Explicit allowlist: empty Origin (non-browser clients) rejected too.
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:8093"
	},
}

func echo(w http.ResponseWriter, r *http.Request) {
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
