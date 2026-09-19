// Package rls implements domain.DocumentRepository through Postgres
// Row Level Security: queries carry no tenant filter, the policy decides
// visibility from transaction-local context.
package rls

import (
	"context"
	"strconv"

	"gorm.io/gorm"

	"go-lab/storage/rls/internal/domain"
	"go-lab/storage/rls/internal/infrastructure/model"
)

// Compile-time check that DocumentRepository implements domain.DocumentRepository.
var _ domain.DocumentRepository = (*DocumentRepository)(nil)

// DocumentRepository relies on the tenant_isolation policy.
// Public (chosen at compose time): mains wire this implementation
// explicitly, so callers name the concrete type.
type DocumentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository builds a DocumentRepository over the given handle.
// It panics on nil: a missing dependency must fail at wiring time.
func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	if db == nil {
		panic("rls: nil gorm.DB")
	}

	return &DocumentRepository{db: db}
}

// ListByTenant sets role and tenant for this transaction only, then runs
// a bare Find. The RLS policy filters rows; the SQL has no WHERE.
func (r *DocumentRepository) ListByTenant(ctx context.Context, tenantID int) ([]domain.Document, error) {
	var models []model.Document

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Exec("SET LOCAL ROLE lab_user").Error
		if err != nil {
			return err
		}

		err = tx.Exec("SET LOCAL app.tenant_id = " + strconv.Itoa(tenantID)).Error
		if err != nil {
			return err
		}

		err = tx.Find(&models).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	docs := make([]domain.Document, 0, len(models))

	for _, m := range models {
		docs = append(docs, m.ToDomain())
	}

	return docs, nil
}
