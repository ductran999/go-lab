// Package domain holds the enterprise entities and repository contracts.
// It has no dependency on any framework, database, or transport detail.
// The entity is a raw struct: JSON mapping lives in infrastructure DTOs.
package domain

import (
	"context"
	"time"
)

// A Todo is the business entity mirrored by the api.todos table
// exposed through PostgREST. TenantID scopes the row to one tenant
// and is enforced by Row Level Security (see docs/02-auth-model.md).
type Todo struct {
	ID       int
	Done     bool
	Task     string
	Due      *time.Time
	TenantID int
}

// TodoRepository abstracts persistence of Todo entities.
// Implementations live in the infrastructure layer.
type TodoRepository interface {
	Create(ctx context.Context, task string, tenantID int) (Todo, error)
	List(ctx context.Context) ([]Todo, error)
	SetDone(ctx context.Context, id int, done bool) (Todo, error)
	Delete(ctx context.Context, id int) error
}
