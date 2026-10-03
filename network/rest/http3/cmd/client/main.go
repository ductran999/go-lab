// Command client hits the HTTP/3 lab: hello then the finite stream.
// GRPC_CA/HTTP3_CA points at the demo CA (default: grpc lab ca.crt).
package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/quic-go/quic-go/http3"

	"github.com/ductran999/shared-pkg/environ"
	"github.com/ductran999/shared-pkg/pretty/display"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "client:", err)

	os.Exit(1)
}

func main() {
	caPEM, err := os.ReadFile(environ.Get("HTTP3_CA", "../../rpc/grpc/certs/ca.crt"))
	if err != nil {
		fail(fmt.Errorf("ca: %w", err))
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caPEM)

	client := &http.Client{
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
		},
	}

	base := "https://localhost:" + environ.Get("PORT", "8108")

	resp, err := client.Get(base + "/hello")
	if err != nil {
		fail(err)
	}

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	show("GET", "/hello", resp.Proto, resp.Status, body)

	resp, err = client.Get(base + "/stream")
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// Streaming read: print each event as it arrives (not ReadAll —
	// that waits for EOF and hides the stream nature).
	fmt.Printf("stream: proto=%s status=%s (live)\n", resp.Proto, resp.Status)

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			fmt.Println("  [dispatch] blank line → event fires")
			continue
		}

		if line[0] == ':' {
			fmt.Println("  [heartbeat]", line)
			continue
		}

		// The server's closing event, not data: stop cleanly on it.
		if line == "event: done" {
			fmt.Println("  [done] server closing event → break")
			break
		}

		fmt.Println("  [data]", line)
	}

	scanErr := scanner.Err()
	if scanErr != nil {
		fail(scanErr)
	}

	fmt.Println("stream ended (FIN received, clean EOF)")
}

func show(method, path, proto, status string, body []byte) {
	var decoded any

	decodeErr := json.Unmarshal(body, &decoded)
	if decodeErr != nil {
		decoded = string(body)
	}

	displayErr := display.PrintJSON(map[string]any{
		"method": method,
		"path":   path,
		"proto":  proto,
		"status": status,
		"body":   decoded,
	})
	if displayErr != nil {
		fail(displayErr)
	}
}
