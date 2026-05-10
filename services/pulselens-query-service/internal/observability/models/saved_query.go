package models

import "time"

type SavedQuery struct {
	ID         string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID   string    `gorm:"index;not null" json:"tenant_id"`
	Name       string    `gorm:"not null" json:"name"`
	QueryType  string    `gorm:"index;not null" json:"query_type"`
	Definition string    `gorm:"type:jsonb;not null" json:"definition"`
	CreatedBy  string    `gorm:"index" json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
