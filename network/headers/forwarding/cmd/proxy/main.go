// Command proxy is one honest hop: it appends the socket peer to
// X-Forwarded-For and forwards everything else untouched.
package main

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func main() {
	target, err := url.Parse("http://" + environ.Get("ORIGIN", "localhost:8103"))
	if err != nil {
		fail(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	addr := ":" + environ.Get("PORT", "8104")

	slog.Info("serving forwarding proxy", "addr", addr, "origin", target.String())

	server := &http.Server{
		Addr:              addr,
		Handler:           proxy,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}
