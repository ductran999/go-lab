// Package domain holds the enterprise entities and repository contracts.
// It has no dependency on any framework, database, or transport detail.
package domain

import (
	"context"
	"time"
)

// A Todo is the business entity mirrored by the api.todos table
// exposed through PostgREST.
type Todo struct {
	ID   int        `json:"id,omitempty"`
	Done bool       `json:"done"`
	Task string     `json:"task"`
	Due  *time.Time `json:"due,omitempty"`
}

// TodoRepository abstracts persistence of Todo entities.
// Implementations live in the infrastructure layer.
type TodoRepository interface {
	Create(ctx context.Context, task string) (Todo, error)
	List(ctx context.Context) ([]Todo, error)
	SetDone(ctx context.Context, id int, done bool) (Todo, error)
	Delete(ctx context.Context, id int) error
}
