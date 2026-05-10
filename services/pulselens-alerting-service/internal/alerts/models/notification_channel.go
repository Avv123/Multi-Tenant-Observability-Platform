package models

import "time"

type NotificationChannel struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	TenantID  string    `gorm:"index;not null" json:"tenant_id"`
	Name      string    `gorm:"not null" json:"name"`
	Type      string    `gorm:"index;not null" json:"type"`
	Config    string    `gorm:"type:jsonb;not null" json:"config"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
