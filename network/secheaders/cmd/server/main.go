// Command server shows the same page twice: /bare with no security
// headers, /hardened with the full set. Diff the responses — the
// body is identical, the protection is all in the headers.
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

const page = `<!doctype html><html><body><h1>same body, different armor</h1></body></html>`

// demoPage loads the same JS twice: bare (executes) vs nosniff (blocked).
const demoPage = `<!doctype html><html><body>
<h1>nosniff on subresources</h1>
<script src="/lib-bare.js"></script>
<script src="/lib-hard.js"></script>
<script>
window.addEventListener("load", () => {
  document.body.insertAdjacentHTML("beforeend",
    "<p>bare script ran: " + (window.bareRan === true) + "</p>" +
    "<p>nosniff script ran: " + (window.hardRan === true) + "</p>");
});
</script></body></html>`

func demo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, demoPage)
}

func libBare(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, `window.bareRan = true;`)
}

func libHard(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, `window.hardRan = true;`)
}

// polyglot looks like HTML but is served as text/plain: a sniffing
// browser executes the script, a nosniff browser shows inert text.
const polyglot = `<html><body><h1>if you see a popup, sniffing executed me</h1><script>alert("sniffed XSS")</script></body></html>`

func bare(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, page)
}

func barePolyglot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, polyglot)
}

func hardenedPolyglot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, polyglot)
}

func hardened(w http.ResponseWriter, r *http.Request) {
	// No sniffing: declared Content-Type is final.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// No framing anywhere: kills clickjacking.
	w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
	// Minimal referrer leakage on outbound navigation.
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	// HTTPS-only for a year, subdomains included. Browsers honor
	// this ONLY over TLS; on plain HTTP it is sent but ignored.
	if r.TLS != nil {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, page)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/bare", bare)
	mux.HandleFunc("/hardened", hardened)
	mux.HandleFunc("/polyglot-bare", barePolyglot)
	mux.HandleFunc("/polyglot-hardened", hardenedPolyglot)
	mux.HandleFunc("/demo", demo)
	mux.HandleFunc("/lib-bare.js", libBare)
	mux.HandleFunc("/lib-hard.js", libHard)

	addr := ":" + environ.Get("PORT", "8105")

	slog.Info("serving secheaders lab", "addr", addr)

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
