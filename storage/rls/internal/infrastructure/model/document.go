// Package model holds persistence models for the lab schema.
// GORM tags and table names live here, never in domain entities.
// Both repository implementations share this model because they
// target the same table.
package model

import (
	"go-lab/storage/rls/internal/domain"
)

// Document maps the lab.documents table for GORM.
type Document struct {
	ID       int    `gorm:"column:id;primaryKey"`
	TenantID int    `gorm:"column:tenant_id"`
	Body     string `gorm:"column:body"`
}

// TableName pins the lab schema explicitly for GORM.
func (Document) TableName() string {
	return "lab.documents"
}

// ToDomain converts the persistence model into a raw domain entity.
func (m Document) ToDomain() domain.Document {
	return domain.Document{
		ID:       m.ID,
		TenantID: m.TenantID,
		Body:     m.Body,
	}
}
