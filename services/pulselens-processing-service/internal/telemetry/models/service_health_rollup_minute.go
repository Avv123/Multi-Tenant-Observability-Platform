package models

import "time"

type ServiceHealthRollupMinute struct {
	ID               string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID         string    `gorm:"uniqueIndex:ux_service_health_rollup_minute;index;not null" json:"tenant_id"`
	ServiceID        string    `gorm:"uniqueIndex:ux_service_health_rollup_minute;index;not null" json:"service_id"`
	Environment      string    `gorm:"uniqueIndex:ux_service_health_rollup_minute;index;not null" json:"environment"`
	ServiceName      string    `gorm:"not null" json:"service_name"`
	BucketStart      time.Time `gorm:"uniqueIndex:ux_service_health_rollup_minute;index;not null" json:"bucket_start"`
	EventCount       int64     `gorm:"not null" json:"event_count"`
	ErrorLogCount    int64     `gorm:"not null" json:"error_log_count"`
	CriticalLogCount int64     `gorm:"not null" json:"critical_log_count"`
	LastEventAt      time.Time `gorm:"not null" json:"last_event_at"`
	LatestMetricAt   time.Time `json:"latest_metric_at"`
	LatestTraceAt    time.Time `json:"latest_trace_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
