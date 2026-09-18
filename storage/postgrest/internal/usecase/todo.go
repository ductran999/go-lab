// Package usecase holds the application business rules.
// It depends only on the domain layer, never on infrastructure details.
package usecase

import (
	"context"
	"errors"
	"strings"

	"go-lab/storage/postgrest/internal/domain"
)

// ErrTaskRequired is returned when a todo task is blank.
var ErrTaskRequired = errors.New("task is required")

// ErrTenantRequired is returned when the tenant id is missing.
// Every todo must belong to a tenant (RLS isolation).
var ErrTenantRequired = errors.New("tenant id is required")

// TodoUseCase orchestrates Todo operations on top of a domain.TodoRepository.
type TodoUseCase struct {
	repo domain.TodoRepository
}

// NewTodoUseCase builds a TodoUseCase with the given repository.
// It panics on a nil repository: a missing dependency is a programmer
// error that must surface at wiring time, not as a nil dereference later.
func NewTodoUseCase(repo domain.TodoRepository) *TodoUseCase {
	if repo == nil {
		panic("usecase: nil TodoRepository")
	}

	return &TodoUseCase{repo: repo}
}

// CreateTodo validates the input, then persists a new todo for the tenant.
func (u *TodoUseCase) CreateTodo(ctx context.Context, task string, tenantID int) (domain.Todo, error) {
	if strings.TrimSpace(task) == "" {
		return domain.Todo{}, ErrTaskRequired
	}

	if tenantID <= 0 {
		return domain.Todo{}, ErrTenantRequired
	}

	return u.repo.Create(ctx, task, tenantID)
}

// ListTodos returns all todos ordered by id.
func (u *TodoUseCase) ListTodos(ctx context.Context) ([]domain.Todo, error) {
	return u.repo.List(ctx)
}

// CompleteTodo marks the todo with the given id as done (or not done).
func (u *TodoUseCase) CompleteTodo(ctx context.Context, id int, done bool) (domain.Todo, error) {
	return u.repo.SetDone(ctx, id, done)
}

// RemoveTodo deletes the todo with the given id.
func (u *TodoUseCase) RemoveTodo(ctx context.Context, id int) error {
	return u.repo.Delete(ctx, id)
}
