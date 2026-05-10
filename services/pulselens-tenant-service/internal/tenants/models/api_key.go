package models

import "time"

type APIKey struct {
	ID         string     `gorm:"primaryKey;type:text" json:"id"`
	TenantID   string     `gorm:"index;not null" json:"tenant_id"`
	ServiceID  string     `gorm:"index;not null" json:"service_id"`
	Name       string     `gorm:"not null" json:"name"`
	KeyPrefix  string     `gorm:"index;not null" json:"key_prefix"`
	KeyHash    string     `gorm:"uniqueIndex;not null" json:"-"`
	Scopes     string     `gorm:"type:jsonb;default:'[]'" json:"scopes"`
	Active     bool       `gorm:"not null;default:true" json:"active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
