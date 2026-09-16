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
	"os"
	"time"

	"go-lab/storage/postgrest/internal/infrastructure/postgrest"
	"go-lab/storage/postgrest/internal/usecase"
)

func baseURL() string {
	if v := os.Getenv("POSTGREST_URL"); v != "" {
		return v
	}

	return "http://localhost:3000"
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "FAILED:", err)

	os.Exit(1)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := postgrest.NewClient(baseURL())
	repo := postgrest.NewTodoRepository(client)
	todos := usecase.NewTodoUseCase(repo)

	fmt.Println("== PostgREST CRUD demo ==")

	created, err := todos.CreateTodo(ctx, "Learn PostgREST with Go")
	if err != nil {
		fail(fmt.Errorf("CREATE: %w", err))
	}

	fmt.Printf("1. CREATED: %+v\n", created)

	list, err := todos.ListTodos(ctx)
	if err != nil {
		fail(fmt.Errorf("LIST: %w", err))
	}

	fmt.Printf("2. LIST (%d todos):\n", len(list))

	for _, t := range list {
		fmt.Printf("   - #%d done=%v task=%q\n", t.ID, t.Done, t.Task)
	}

	updated, err := todos.CompleteTodo(ctx, created.ID, true)
	if err != nil {
		fail(fmt.Errorf("PATCH: %w", err))
	}

	fmt.Printf("3. UPDATED: %+v\n", updated)

	err = todos.RemoveTodo(ctx, created.ID)
	if err != nil {
		fail(fmt.Errorf("DELETE: %w", err))
	}

	fmt.Printf("4. DELETED todo #%d\n", created.ID)

	remaining, err := todos.ListTodos(ctx)
	if err != nil {
		fail(fmt.Errorf("LIST after delete: %w", err))
	}

	fmt.Printf("5. REMAINING (%d todos)\n", len(remaining))
}
