package models

import "time"

type RetryEvent struct {
	ID            string     `gorm:"primaryKey;type:text" json:"id"`
	EventID       string     `gorm:"index;not null" json:"event_id"`
	TenantID      string     `gorm:"index;not null" json:"tenant_id"`
	ServiceID     string     `gorm:"index;not null" json:"service_id"`
	EventType     string     `gorm:"index;not null" json:"event_type"`
	Payload       string     `gorm:"type:jsonb" json:"payload"`
	RetryCount    int        `json:"retry_count"`
	Status        string     `gorm:"index;not null" json:"status"`
	NextAttemptAt time.Time  `gorm:"index;not null" json:"next_attempt_at"`
	DispatchedAt  *time.Time `json:"dispatched_at"`
	LastError     string     `gorm:"type:text" json:"last_error"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
