package models

import "time"

type TraceLatencyRollupMinute struct {
	ID              string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID        string    `gorm:"uniqueIndex:ux_trace_latency_rollup_minute;index;not null" json:"tenant_id"`
	ServiceID       string    `gorm:"uniqueIndex:ux_trace_latency_rollup_minute;index;not null" json:"service_id"`
	ServiceName     string    `gorm:"not null" json:"service_name"`
	Environment     string    `gorm:"uniqueIndex:ux_trace_latency_rollup_minute;index;not null" json:"environment"`
	Operation       string    `gorm:"uniqueIndex:ux_trace_latency_rollup_minute;index;not null" json:"operation"`
	BucketStart     time.Time `gorm:"uniqueIndex:ux_trace_latency_rollup_minute;index;not null" json:"bucket_start"`
	SpanCount       int64     `gorm:"not null" json:"span_count"`
	ErrorCount      int64     `gorm:"not null" json:"error_count"`
	TotalDurationMS int64     `gorm:"not null" json:"total_duration_ms"`
	MaxDurationMS   int64     `gorm:"not null" json:"max_duration_ms"`
	LastEventAt     time.Time `gorm:"not null" json:"last_event_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
