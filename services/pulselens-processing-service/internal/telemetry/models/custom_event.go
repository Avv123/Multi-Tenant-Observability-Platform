package models

import "time"

type CustomEvent struct {
	EventID     string    `gorm:"primaryKey;type:text" json:"event_id"`
	TenantID    string    `gorm:"index;not null" json:"tenant_id"`
	ServiceID   string    `gorm:"index;not null" json:"service_id"`
	ServiceName string    `gorm:"index;not null" json:"service_name"`
	Environment string    `gorm:"index;not null" json:"environment"`
	ShardID     int       `gorm:"index;default:0" json:"shard_id"`
	Payload     string    `gorm:"type:jsonb" json:"payload"`
	OccurredAt  time.Time `gorm:"index" json:"occurred_at"`
	ReceivedAt  time.Time `gorm:"index" json:"received_at"`
	CreatedAt   time.Time `json:"created_at"`
}
