package models

import "time"

type ArchiveRecord struct {
	ID            string    `gorm:"primaryKey;type:text" json:"id"`
	EventID       string    `gorm:"index;not null" json:"event_id"`
	TenantID      string    `gorm:"index;not null" json:"tenant_id"`
	ServiceID     string    `gorm:"index;not null" json:"service_id"`
	EventType     string    `gorm:"index;not null" json:"event_type"`
	Payload       string    `gorm:"type:jsonb;not null" json:"payload"`
	ArchiveBucket string    `gorm:"type:text" json:"archive_bucket"`
	ArchiveKey    string    `gorm:"type:text;index" json:"archive_key"`
	ArchivePath   string    `gorm:"type:text;not null" json:"archive_path"`
	OccurredAt    time.Time `gorm:"index;not null" json:"occurred_at"`
	CreatedAt     time.Time `json:"created_at"`
}
