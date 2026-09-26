// Command server demonstrates CORS mechanics with a hand-rolled
// middleware (no library, every header visible): /api/* answers
// preflights, /plain/* does not. Compare with curl.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ductran999/shared-pkg/environ"

	"go-lab/network/http/cors/internal/session"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

// cors answers CORS preflights and tags real responses.
// Hand-rolled on purpose: each header below is one CORS concept.
// Global middleware gated on /api/*: it must run even for OPTIONS,
// which matches no GET/POST route (group middleware would not run).
func cors(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()

			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()

			return
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Vary", "Origin")

		if c.Request.Method != http.MethodOptions {
			c.Next()

			return
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,Prefer")
		c.Header("Access-Control-Max-Age", "86400")
		c.AbortWithStatus(http.StatusNoContent)
	}
}

// corsCredentials is the credentials-mode variant: it echoes the
// request Origin (never "*") plus Access-Control-Allow-Credentials.
// Browsers reject "*" + credentials outright, so echo is mandatory.
// Vary: Origin is load-bearing here: each origin caches its own ACAO.
func corsCredentials() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/cred/") {
			c.Next()

			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()

			return
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Vary", "Origin")

		if c.Request.Method != http.MethodOptions {
			c.Next()

			return
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Max-Age", "86400")
		c.AbortWithStatus(http.StatusNoContent)
	}
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// trace sets two custom response headers but exposes only one.
// JS reads X-Request-Id fine and gets null for X-Internal-Flag:
// safelisted headers (Cache-Control, Content-Type...) plus exposed
// ones are visible, everything else is hidden from frontend code.
func trace(c *gin.Context) {
	c.Header("X-Request-Id", "req-123")
	c.Header("X-Internal-Flag", "secret")
	c.Header("Access-Control-Expose-Headers", "X-Request-Id")
	c.JSON(http.StatusOK, gin.H{"message": "traced"})
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(cors("http://localhost:8091"))
	r.Use(corsCredentials())

	api := r.Group("/api")
	api.GET("/ping", ping)
	api.POST("/ping", ping)
	api.PUT("/ping", ping)
	api.GET("/trace", trace)

	cred := r.Group("/cred")
	cred.POST("/login", session.Login)
	cred.GET("/me", session.Me)

	// No middleware: preflights 404 here.
	r.GET("/plain/ping", ping)

	addr := ":" + environ.Get("PORT", "8090")

	slog.Info("serving CORS lab", "addr", addr)

	err := r.Run(addr)
	if err != nil {
		fail(err)
	}
}
