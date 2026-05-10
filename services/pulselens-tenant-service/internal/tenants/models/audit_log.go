package models

import "time"

type AuditLog struct {
	ID           string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID     string    `gorm:"index;not null" json:"tenant_id"`
	ActorUserID  string    `gorm:"index" json:"actor_user_id"`
	ActorType    string    `gorm:"index;not null" json:"actor_type"`
	Action       string    `gorm:"index;not null" json:"action"`
	ResourceType string    `gorm:"index;not null" json:"resource_type"`
	ResourceID   string    `gorm:"index;not null" json:"resource_id"`
	Payload      string    `gorm:"type:jsonb;default:'{}'" json:"payload"`
	CreatedAt    time.Time `json:"created_at"`
}
