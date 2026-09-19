// Command postgrest-demo runs one CRUD cycle against PostgREST.
//
// It is the composition root: it wires the infrastructure implementation
// (PostgREST HTTP repository) into the use case layer and drives a demo flow.
// Swap the repository constructor here to target a different data source
// without touching domain or use case code.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go-lab/storage/postgrest/internal/infrastructure/postgrest"
	"go-lab/storage/postgrest/internal/usecase"

	"github.com/ductran999/shared-pkg/environ"
)

func baseURL() string {
	return environ.Get("POSTGREST_URL", "http://localhost:3000")
}

// demoTenantID scopes demo rows to tenant 1. POSTGREST_TOKEN optionally
// authenticates as that tenant (see docs/02-auth-model.md); empty runs
// anonymous via the lab open policy.
const demoTenantID = 1

func fail(err error) {
	slog.Error("demo failed", "error", err)

	os.Exit(1)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := postgrest.NewClient(baseURL(), environ.Get("POSTGREST_TOKEN", ""))
	repo := postgrest.NewTodoRepository(client)
	todos := usecase.NewTodoUseCase(repo)

	slog.Info("starting PostgREST CRUD demo", "tenant", demoTenantID)

	created, err := todos.CreateTodo(ctx, "Learn PostgREST with Go", demoTenantID)
	if err != nil {
		fail(fmt.Errorf("CREATE: %w", err))
	}

	slog.Info("todo created", "id", created.ID, "done", created.Done, "task", created.Task, "tenant", created.TenantID)

	list, err := todos.ListTodos(ctx)
	if err != nil {
		fail(fmt.Errorf("LIST: %w", err))
	}

	slog.Info("todos listed", "count", len(list))

	for _, t := range list {
		slog.Info("todo", "id", t.ID, "done", t.Done, "task", t.Task)
	}

	updated, err := todos.CompleteTodo(ctx, created.ID, true)
	if err != nil {
		fail(fmt.Errorf("PATCH: %w", err))
	}

	slog.Info("todo updated", "id", updated.ID, "done", updated.Done, "task", updated.Task)

	err = todos.RemoveTodo(ctx, created.ID)
	if err != nil {
		fail(fmt.Errorf("DELETE: %w", err))
	}

	slog.Info("todo deleted", "id", created.ID)

	remaining, err := todos.ListTodos(ctx)
	if err != nil {
		fail(fmt.Errorf("LIST after delete: %w", err))
	}

	slog.Info("demo finished", "remaining", len(remaining))
}
