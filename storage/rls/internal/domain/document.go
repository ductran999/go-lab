// Package domain holds the Document entity and the repository contract.
// The entity is a raw struct: no JSON, GORM, or transport tags.
// Mapping to outer-layer representations lives outside this package.
package domain

import "context"

// Document is a tenant-scoped record in lab.documents.
type Document struct {
	ID       int
	TenantID int
	Body     string
}

// DocumentRepository lists documents visible to one tenant.
// How visibility is enforced is an infrastructure decision.
type DocumentRepository interface {
	ListByTenant(ctx context.Context, tenantID int) ([]Document, error)
}
