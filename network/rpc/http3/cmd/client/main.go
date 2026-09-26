// Command client hits the HTTP/3 lab: hello then the finite stream.
// GRPC_CA/HTTP3_CA points at the demo CA (default: grpc lab ca.crt).
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/quic-go/quic-go/http3"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "client:", err)

	os.Exit(1)
}

func main() {
	caPEM, err := os.ReadFile(environ.Get("HTTP3_CA", "../grpc/certs/ca.crt"))
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

	fmt.Printf("hello: proto=%s body=%s", resp.Proto, body)

	resp, err = client.Get(base + "/stream")
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fail(err)
	}

	fmt.Printf("stream: proto=%s\n%s", resp.Proto, body)
}
