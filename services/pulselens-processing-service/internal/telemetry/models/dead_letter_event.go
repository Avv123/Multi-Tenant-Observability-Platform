package models

import "time"

type DeadLetterEvent struct {
	EventID    string    `gorm:"primaryKey;type:text" json:"event_id"`
	TenantID   string    `gorm:"index;not null" json:"tenant_id"`
	ServiceID  string    `gorm:"index;not null" json:"service_id"`
	EventType  string    `gorm:"index;not null" json:"event_type"`
	Reason     string    `gorm:"type:text" json:"reason"`
	Payload    string    `gorm:"type:jsonb" json:"payload"`
	RetryCount int       `json:"retry_count"`
	ReceivedAt time.Time `gorm:"index" json:"received_at"`
	CreatedAt  time.Time `json:"created_at"`
}
