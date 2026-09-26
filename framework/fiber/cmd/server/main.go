// Fiber twin of the mux lesson: same Todos JSON API, fiber dialect
// (fasthttp under the hood — NOT net/http: no http.Server, no hijack
// compat). Compare: c.Params/c.JSON like gin/echo, app.Use, 404 via
// catch-all. Run: go run . (curl :8140/todos).
package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

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
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})
	app.Get("/todos", func(c *fiber.Ctx) error {
		return c.JSON(todos)
	})
	app.Get("/todos/:id", func(c *fiber.Ctx) error {
		for _, td := range todos {
			if td.ID == c.Params("id") {
				return c.JSON(td)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(map[string]string{"message": "not found"})
	})

	addr := ":" + environ.Get("PORT", "8140")

	slog.Info("fiber twin", "addr", addr)

	err := app.Listen(addr)
	if err != nil {
		fail(err)
	}
}
