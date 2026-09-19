// Package usecase holds application business rules.
// It depends only on domain: any DocumentRepository implementation
// (manual scoping or RLS) plugs in unchanged.
package usecase

import (
	"context"

	"go-lab/storage/rls/internal/domain"
)

// TenantDocs groups one tenant with its visible documents.
type TenantDocs struct {
	TenantID int
	Docs     []domain.Document
}

// Reporter builds per-tenant document reports.
type Reporter struct {
	repo domain.DocumentRepository
}

// NewReporter builds a Reporter over the given repository.
// It panics on nil: a missing dependency must fail at wiring time.
func NewReporter(repo domain.DocumentRepository) *Reporter {
	if repo == nil {
		panic("usecase: nil DocumentRepository")
	}

	return &Reporter{repo: repo}
}

// BuildReport lists documents for each tenant in order.
// Visibility rules live in the repository implementation,
// so this flow is identical for every enforcement style.
func (u *Reporter) BuildReport(ctx context.Context, tenants []int) ([]TenantDocs, error) {
	report := make([]TenantDocs, 0, len(tenants))

	for _, tenantID := range tenants {
		docs, err := u.repo.ListByTenant(ctx, tenantID)
		if err != nil {
			return nil, err
		}

		report = append(report, TenantDocs{TenantID: tenantID, Docs: docs})
	}

	return report, nil
}
