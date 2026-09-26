// Gin twin of the mux lesson: same Todos JSON API, gin dialect.
// Compare: c.JSON/c.Param replace writeJSON/PathValue, middleware via
// Use(), 404 via NoRoute. Run: go run . (curl :8120/todos).
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("lesson failed", "error", err)

	os.Exit(1)
}

type Todo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

var todos = []Todo{
	{ID: "1", Description: "learn mux"},
	{ID: "2", Description: "write todo app"},
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/todos", func(c *gin.Context) {
		c.JSON(http.StatusOK, todos)
	})
	r.GET("/todos/:id", func(c *gin.Context) {
		for _, td := range todos {
			if td.ID == c.Param("id") {
				c.JSON(http.StatusOK, td)

				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	})

	addr := ":" + environ.Get("PORT", "8120")

	slog.Info("gin twin", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}
