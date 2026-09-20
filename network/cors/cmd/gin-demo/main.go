// Command gin-demo is the same lab API as cmd/server but with the
// community cors middleware instead of the hand-rolled one.
// Compare behaviors with identical curl commands; configs mirror
// each other line by line.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/ductran999/shared-pkg/environ"

	"go-lab/network/cors/internal/session"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// trace mirrors cmd/server: one exposed header, one hidden.
func trace(c *gin.Context) {
	c.Header("X-Request-Id", "req-123")
	c.Header("X-Internal-Flag", "secret")
	c.Header("Access-Control-Expose-Headers", "X-Request-Id")
	c.JSON(http.StatusOK, gin.H{"message": "traced"})
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Same lesson as the hand-rolled server: the library middleware must
	// run globally (unregistered OPTIONS matches no route, so group
	// middleware would never fire), gated on paths that opt into CORS.
	libCors := cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8091"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Prefer", "Origin"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	})

	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/cred/") {
			c.Next()

			return
		}

		libCors(c)
	})

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

	addr := ":" + environ.Get("PORT", "8092")

	slog.Info("serving gin-cors lab", "addr", addr)

	err := r.Run(addr)
	if err != nil {
		fail(err)
	}
}
