// Command origin reports what it sees: socket peer, forwarding
// headers, and the trusted verdict on the real client IP.
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go-lab/network/forwarding/internal/forward"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func whoami(w http.ResponseWriter, r *http.Request) {
	// Only our own proxy (localhost) may vouch for clients.
	trusted := []string{"127.0.0.1", "::1"}

	if extra := environ.Get("TRUSTED_PROXIES", ""); extra != "" {
		trusted = append(trusted, extra)
	}

	out := map[string]string{
		"remote_addr":     r.RemoteAddr,
		"x_forwarded_for": r.Header.Get("X-Forwarded-For"),
		"forwarded":       r.Header.Get("Forwarded"),
		"real_ip":         forward.RealIP(r, trusted),
	}

	peer, _, _ := net.SplitHostPort(r.RemoteAddr)

	switch {
	case out["x_forwarded_for"] != "" && out["real_ip"] != peer:
		out["trusted_decision"] = "XFF honored (peer is a known proxy)"
	case out["x_forwarded_for"] != "":
		out["trusted_decision"] = "XFF present but peer untrusted — socket peer used"
	default:
		out["trusted_decision"] = "no forwarding headers — socket peer is the truth"
	}

	// Blocklist applies to the trusted verdict, never raw headers.
	for b := range strings.SplitSeq(environ.Get("BLOCKED_IPS", ""), ",") {
		if b = strings.TrimSpace(b); b != "" && b == out["real_ip"] {
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprint(w, `{"error":"blocked"}`)

			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/whoami", whoami)

	addr := ":" + environ.Get("PORT", "8103")

	slog.Info("serving forwarding origin", "addr", addr)

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
