package models

import "time"

type Service struct {
	ID          string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID    string    `gorm:"index;not null" json:"tenant_id"`
	Name        string    `gorm:"not null" json:"name"`
	Environment string    `gorm:"not null" json:"environment"`
	Tags        string    `gorm:"type:jsonb;default:'{}'" json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
