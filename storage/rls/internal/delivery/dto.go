package delivery

import (
	"go-lab/storage/rls/internal/domain"
)

// DocumentDTO is the JSON representation of a domain Document.
type DocumentDTO struct {
	ID       int    `json:"id"`
	TenantID int    `json:"tenant_id"`
	Body     string `json:"body"`
}

// ToDTOs maps domain entities to display DTOs, never nil so
// JSON responses encode [] instead of null.
func ToDTOs(docs []domain.Document) []DocumentDTO {
	dtos := make([]DocumentDTO, 0, len(docs))

	for _, d := range docs {
		dtos = append(dtos, DocumentDTO{
			ID:       d.ID,
			TenantID: d.TenantID,
			Body:     d.Body,
		})
	}

	return dtos
}
