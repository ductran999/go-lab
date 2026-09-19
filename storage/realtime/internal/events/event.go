// Package events holds the app.events row shape shared by the writer
// endpoint and the SSE stream. PoC pragmatism: one struct carries both
// GORM mapping and JSON tags; split on growth like the rls lab did.
package events

import (
	"encoding/json"
	"time"
)

// Event is one row of app.events and one SSE data frame.
type Event struct {
	ID        int             `gorm:"column:id;primaryKey" json:"id"`
	TenantID  int             `gorm:"column:tenant_id" json:"tenant_id"`
	Kind      string          `gorm:"column:kind" json:"kind"`
	Payload   json.RawMessage `gorm:"column:payload;type:jsonb" json:"payload"`
	CreatedAt time.Time       `gorm:"column:created_at" json:"created_at"`
}

// TableName pins the app schema explicitly for GORM.
func (Event) TableName() string {
	return "app.events"
}
