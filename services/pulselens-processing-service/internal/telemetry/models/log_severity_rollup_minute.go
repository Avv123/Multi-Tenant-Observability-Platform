package models

import "time"

type LogSeverityRollupMinute struct {
	ID          string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID    string    `gorm:"uniqueIndex:ux_log_severity_rollup_minute;index;not null" json:"tenant_id"`
	ServiceID   string    `gorm:"uniqueIndex:ux_log_severity_rollup_minute;index;not null" json:"service_id"`
	ServiceName string    `gorm:"not null" json:"service_name"`
	Environment string    `gorm:"uniqueIndex:ux_log_severity_rollup_minute;index;not null" json:"environment"`
	Severity    string    `gorm:"uniqueIndex:ux_log_severity_rollup_minute;index;not null" json:"severity"`
	BucketStart time.Time `gorm:"uniqueIndex:ux_log_severity_rollup_minute;index;not null" json:"bucket_start"`
	EventCount  int64     `gorm:"not null" json:"event_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
