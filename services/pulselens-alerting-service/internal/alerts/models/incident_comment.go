package models

import "time"

type IncidentComment struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	IncidentID string    `gorm:"index;not null" json:"incident_id"`
	TenantID   string    `gorm:"index;not null" json:"tenant_id"`
	AuthorID   string    `gorm:"index" json:"author_id"`
	Body       string    `gorm:"type:text;not null" json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}
