// Package forward decides the real client IP: trust X-Forwarded-For
// only from known proxies, otherwise the socket peer is the truth.
package forward

import (
	"net"
	"net/http"
	"strings"
)

// RealIP returns the client IP. trustedProxies lists IPs allowed to
// append X-Forwarded-For; anything else is client-controlled noise.
func RealIP(r *http.Request, trustedProxies []string) string {
	peer, _, _ := net.SplitHostPort(r.RemoteAddr)

	trusted := false

	for _, t := range trustedProxies {
		if peer == t || peer == "::1" && t == "127.0.0.1" {
			trusted = true

			break
		}
	}

	if !trusted {
		return peer
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Leftmost = original client; each proxy appends to the right.
		if first := strings.TrimSpace(strings.Split(xff, ",")[0]); first != "" {
			return first
		}
	}

	return peer
}
