package models

import "time"

type User struct {
	ID           string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID     string    `gorm:"index;not null" json:"tenant_id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"not null" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
