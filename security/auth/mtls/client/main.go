package main

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// keyPath resolves client.key next to this source file, so the demo
// runs from any working directory. Keys stay out of git (*.key ignored).
func keyPath() string {
	_, thisFile, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(thisFile), "client.key")
}

//go:embed ca.crt
var caCertPEM []byte // Public CA to verify the Server

//go:embed client.crt
var clientCertPEM []byte // Public Client Cert to present to the Server

func main() {
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertPEM); !ok {
		log.Fatal("Failed to parse embedded CA certificate")
	}

	clientKeyPEM, err := os.ReadFile(keyPath())
	if err != nil {
		log.Fatal("Error reading client.key from disk: ", err)
	}

	clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	if err != nil {
		log.Fatal("Error parsing client key pair: ", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{clientCert},
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: tr}

	resp, err := httpClient.Get("https://localhost:8443/secure-api")
	if err != nil {
		log.Fatal("Connection failed: ", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Server Response:", string(body))
}
