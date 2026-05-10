package models

import "time"

type UsageCounter struct {
	ID         string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID   string    `gorm:"uniqueIndex:ux_usage_counter;not null" json:"tenant_id"`
	ServiceID  string    `gorm:"uniqueIndex:ux_usage_counter;not null" json:"service_id"`
	SignalType string    `gorm:"uniqueIndex:ux_usage_counter;not null" json:"signal_type"`
	UsageDate  string    `gorm:"uniqueIndex:ux_usage_counter;not null" json:"usage_date"`
	EventCount int64     `gorm:"not null" json:"event_count"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedAt  time.Time `json:"created_at"`
}
