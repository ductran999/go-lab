// Echo twin of the mux lesson: same Todos JSON API, echo dialect.
// Compare: c.JSON/c.Param like gin, middleware via Use, HTTPErrorHandler
// for 404 JSON. Run: go run . (curl :8130/todos).
package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

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
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError

		var he *echo.HTTPError
		if errors.As(err, &he) {
			code = he.Code
		}

		if !c.Response().Committed {
			_ = c.JSON(code, map[string]string{"message": "not found"})
		}
	}

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/todos", func(c echo.Context) error {
		return c.JSON(http.StatusOK, todos)
	})
	e.GET("/todos/:id", func(c echo.Context) error {
		for _, td := range todos {
			if td.ID == c.Param("id") {
				return c.JSON(http.StatusOK, td)
			}
		}

		return echo.ErrNotFound
	})

	addr := ":" + environ.Get("PORT", "8130")

	slog.Info("echo twin", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           e,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}
