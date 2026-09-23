package main

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

//go:embed ca.crt
var caCertPEM []byte

//go:embed server.crt
var serverCertPEM []byte

// keyPath resolves server.key next to this source file, so the demo
// runs from any working directory. Keys stay out of git (*.key ignored).
func keyPath() string {
	_, thisFile, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(thisFile), "server.key")
}

func main() {
	serverKeyPEM, err := os.ReadFile(keyPath())
	if err != nil {
		log.Fatal("Error reading server.key from disk: ", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)

	serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		log.Fatal("Error parsing embedded server key pair:", err)
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		Certificates: []tls.Certificate{serverCert},
	}

	r := gin.Default()

	r.GET("/secure-api", func(c *gin.Context) {
		if c.Request.TLS != nil && len(c.Request.TLS.PeerCertificates) > 0 {
			clientCert := c.Request.TLS.PeerCertificates[0]
			c.JSON(http.StatusOK, gin.H{
				"message":     "mTLS Handshake Successful!",
				"client_name": clientCert.Subject.CommonName,
			})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: TLS connection required"})
	})

	server := &http.Server{
		Addr:      ":8443",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	fmt.Println("mTLS HTTPS Server is running on https://localhost:8443")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
