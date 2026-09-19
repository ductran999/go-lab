package postgrest

import (
	"time"

	"go-lab/storage/postgrest/internal/domain"
)

// todoDTO is the wire representation of api.todos.
// JSON tags live here (infrastructure layer), never on domain entities.
type todoDTO struct {
	ID       int        `json:"id,omitempty"`
	Done     bool       `json:"done"`
	Task     string     `json:"task"`
	Due      *time.Time `json:"due,omitempty"`
	TenantID int        `json:"tenant_id"`
}

// toDomain converts the wire DTO into a raw domain entity.
func (dto todoDTO) toDomain() domain.Todo {
	return domain.Todo{
		ID:       dto.ID,
		Done:     dto.Done,
		Task:     dto.Task,
		Due:      dto.Due,
		TenantID: dto.TenantID,
	}
}
