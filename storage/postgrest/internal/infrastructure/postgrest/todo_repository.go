package postgrest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"go-lab/storage/postgrest/internal/domain"
)

// ErrEmptyRepresentation is returned when PostgREST answers without a representation.
var ErrEmptyRepresentation = errors.New("postgrest returned empty representation")

// ErrTodoNotFound is returned when no todo matches the requested id.
var ErrTodoNotFound = errors.New("todo not found")

// Compile-time check that TodoRepository implements domain.TodoRepository.
var _ domain.TodoRepository = (*TodoRepository)(nil)

// TodoRepository persists domain.Todo entities through PostgREST.
type TodoRepository struct {
	client *Client
}

// NewTodoRepository builds a TodoRepository using the given client.
// It panics on a nil client: a missing dependency is a programmer
// error that must surface at wiring time, not as a nil dereference later.
func NewTodoRepository(client *Client) *TodoRepository {
	if client == nil {
		panic("postgrest: nil Client")
	}

	return &TodoRepository{client: client}
}

// Create inserts a new todo for the tenant and returns its representation.
func (r *TodoRepository) Create(ctx context.Context, task string, tenantID int) (domain.Todo, error) {
	data, err := r.client.Do(ctx, http.MethodPost, "/todos", nil,
		map[string]any{"task": task, "tenant_id": tenantID}, "return=representation")
	if err != nil {
		return domain.Todo{}, err
	}

	todos, err := decodeTodos(data)
	if err != nil {
		return domain.Todo{}, err
	}

	if len(todos) == 0 {
		return domain.Todo{}, ErrEmptyRepresentation
	}

	return todos[0], nil
}

// List returns all todos ordered by id.
func (r *TodoRepository) List(ctx context.Context) ([]domain.Todo, error) {
	q := url.Values{"select": {"*"}, "order": {"id"}}

	data, err := r.client.Do(ctx, http.MethodGet, "/todos", q, nil, "")
	if err != nil {
		return nil, err
	}

	todos, err := decodeTodos(data)
	if err != nil {
		return nil, err
	}

	if todos == nil {
		todos = []domain.Todo{}
	}

	return todos, nil
}

// SetDone flips the done flag of one todo and returns its representation.
func (r *TodoRepository) SetDone(ctx context.Context, id int, done bool) (domain.Todo, error) {
	q := url.Values{"id": {fmt.Sprintf("eq.%d", id)}}

	data, err := r.client.Do(ctx, http.MethodPatch, "/todos", q,
		map[string]any{"done": done}, "return=representation")
	if err != nil {
		return domain.Todo{}, err
	}

	todos, err := decodeTodos(data)
	if err != nil {
		return domain.Todo{}, err
	}

	if len(todos) == 0 {
		return domain.Todo{}, fmt.Errorf("todo id=%d: %w", id, ErrTodoNotFound)
	}

	return todos[0], nil
}

// Delete removes the todo with the given id.
func (r *TodoRepository) Delete(ctx context.Context, id int) error {
	q := url.Values{"id": {fmt.Sprintf("eq.%d", id)}}

	_, err := r.client.Do(ctx, http.MethodDelete, "/todos", q, nil, "")

	return err
}

func decodeTodos(data []byte) ([]domain.Todo, error) {
	var dtos []todoDTO

	err := json.Unmarshal(data, &dtos)
	if err != nil {
		return nil, fmt.Errorf("decode todos: %w", err)
	}

	todos := make([]domain.Todo, 0, len(dtos))

	for _, dto := range dtos {
		todos = append(todos, dto.toDomain())
	}

	return todos, nil
}
