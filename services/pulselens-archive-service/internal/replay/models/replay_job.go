package models

import "time"

type ReplayJob struct {
	ID           string     `gorm:"primaryKey;type:text" json:"id"`
	TenantID     string     `gorm:"index;not null" json:"tenant_id"`
	ServiceID    string     `gorm:"index" json:"service_id"`
	EventType    string     `gorm:"index" json:"event_type"`
	StartTime    time.Time  `gorm:"index;not null" json:"start_time"`
	EndTime      time.Time  `gorm:"index;not null" json:"end_time"`
	Status       string     `gorm:"index;not null" json:"status"`
	RequestedBy  string     `json:"requested_by"`
	ReplayCount  int64      `json:"replay_count"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
