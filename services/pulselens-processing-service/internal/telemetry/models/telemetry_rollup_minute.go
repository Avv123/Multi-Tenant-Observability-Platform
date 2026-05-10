package models

import "time"

type TelemetryRollupMinute struct {
	ID          string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID    string    `gorm:"uniqueIndex:ux_telemetry_rollup_minute;index;not null" json:"tenant_id"`
	ServiceID   string    `gorm:"uniqueIndex:ux_telemetry_rollup_minute;index;not null" json:"service_id"`
	Environment string    `gorm:"uniqueIndex:ux_telemetry_rollup_minute;index;not null" json:"environment"`
	EventType   string    `gorm:"uniqueIndex:ux_telemetry_rollup_minute;index;not null" json:"event_type"`
	BucketStart time.Time `gorm:"uniqueIndex:ux_telemetry_rollup_minute;index;not null" json:"bucket_start"`
	EventCount  int64     `gorm:"not null" json:"event_count"`
	LastEventAt time.Time `gorm:"not null" json:"last_event_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
