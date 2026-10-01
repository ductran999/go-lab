// Command client talks REST to the H2C lab with prior knowledge:
// no upgrade dance, straight into HTTP/2 frames.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ductran999/shared-pkg/environ"
	"github.com/ductran999/shared-pkg/pretty/display"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "client:", err)

	os.Exit(1)
}

func main() {
	port := environ.Get("PORT", "8109")

	// TLS_CA set → https with verification (same demo CA as grpc lab).
	// Unset → h2c prior knowledge.
	var base string

	var client *http.Client

	if ca := environ.Get("TLS_CA", ""); ca != "" {
		//nolint // lab client: CA path is explicit operator config.
		caPEM, err := os.ReadFile(ca)
		if err != nil {
			fail(err)
		}

		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caPEM)

		base = "https://localhost:" + port
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:   &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
				ForceAttemptHTTP2: true,
			},
		}
	} else {
		base = "http://localhost:" + port

		// Unencrypted HTTP/2, stdlib way (Go 1.24+): prior knowledge,
		// no upgrade dance, no x/net needed.
		tr := &http.Transport{}
		tr.Protocols = &http.Protocols{}
		tr.Protocols.SetHTTP1(true)
		tr.Protocols.SetUnencryptedHTTP2(true)

		client = &http.Client{Transport: tr}
	}

	created, err := client.Post(base+"/todos", "application/json", strings.NewReader(`{"task":"write lab"}`))
	if err != nil {
		fail(err)
	}

	body, _ := io.ReadAll(created.Body)
	_ = created.Body.Close()

	show("POST", created.Proto, created.Status, body)

	resp, err := client.Get(base + "/todos")
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

	show("GET", resp.Proto, resp.Status, body)
}

func show(method, proto, status string, body []byte) {
	var decoded any

	decodeErr := json.Unmarshal(body, &decoded)
	if decodeErr != nil {
		decoded = string(body)
	}

	displayErr := display.PrintJSON(map[string]any{
		"method": method,
		"proto":  proto,
		"status": status,
		"body":   decoded,
	})
	if displayErr != nil {
		fail(displayErr)
	}
}
