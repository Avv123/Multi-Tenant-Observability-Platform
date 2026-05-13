package models

import "time"

type IncidentEvent struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	IncidentID string    `gorm:"index;not null" json:"incident_id"`
	TenantID   string    `gorm:"index;not null" json:"tenant_id"`
	EventType  string    `gorm:"index;not null" json:"event_type"`
	ActorID    string    `gorm:"index" json:"actor_id"`
	Summary    string    `gorm:"type:text;not null" json:"summary"`
	Metadata   string    `gorm:"type:jsonb;not null" json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}
