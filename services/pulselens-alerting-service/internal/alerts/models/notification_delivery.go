package models

import "time"

type NotificationDelivery struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	TenantID     string     `gorm:"index;not null" json:"tenant_id"`
	IncidentID   string     `gorm:"index;not null" json:"incident_id"`
	ChannelID    string     `gorm:"index;not null" json:"channel_id"`
	EventType    string     `gorm:"index;not null" json:"event_type"`
	Status       string     `gorm:"index;not null" json:"status"`
	AttemptCount int        `gorm:"not null;default:0" json:"attempt_count"`
	Payload      string     `gorm:"type:jsonb;not null" json:"payload"`
	Response     string     `gorm:"type:text" json:"response"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
