// Package scoped implements domain.DocumentRepository by adding an
// explicit tenant filter to every query. Correctness depends on no
// query forgetting the Where clause.
package scoped

import (
	"context"

	"gorm.io/gorm"

	"go-lab/storage/rls/internal/domain"
	"go-lab/storage/rls/internal/infrastructure/model"
)

// Compile-time check that DocumentRepository implements domain.DocumentRepository.
var _ domain.DocumentRepository = (*DocumentRepository)(nil)

// DocumentRepository filters by tenant in application code.
// Exported (concrete return): one implementation may satisfy several
// interfaces; callers bind it to whichever narrow contract they need.
type DocumentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository builds a DocumentRepository over the given handle.
// It panics on nil: a missing dependency must fail at wiring time.
func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	if db == nil {
		panic("scoped: nil gorm.DB")
	}

	return &DocumentRepository{db: db}
}

// ListByTenant returns documents with an explicit WHERE on tenant_id.
// Forget this clause anywhere and tenants leak into each other.
func (r *DocumentRepository) ListByTenant(ctx context.Context, tenantID int) ([]domain.Document, error) {
	var models []model.Document

	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	docs := make([]domain.Document, 0, len(models))

	for _, m := range models {
		docs = append(docs, m.ToDomain())
	}

	return docs, nil
}
