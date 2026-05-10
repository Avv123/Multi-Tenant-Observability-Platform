package models

import "time"

type MetricRollupMinute struct {
	ID          string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID    string    `gorm:"uniqueIndex:ux_metric_rollup_minute;index;not null" json:"tenant_id"`
	ServiceID   string    `gorm:"uniqueIndex:ux_metric_rollup_minute;index;not null" json:"service_id"`
	Environment string    `gorm:"uniqueIndex:ux_metric_rollup_minute;index;not null" json:"environment"`
	MetricName  string    `gorm:"uniqueIndex:ux_metric_rollup_minute;index;not null" json:"metric_name"`
	BucketStart time.Time `gorm:"uniqueIndex:ux_metric_rollup_minute;index;not null" json:"bucket_start"`
	SampleCount int64     `gorm:"not null" json:"sample_count"`
	SumValue    float64   `gorm:"not null" json:"sum_value"`
	MinValue    float64   `gorm:"not null" json:"min_value"`
	MaxValue    float64   `gorm:"not null" json:"max_value"`
	LastValue   float64   `gorm:"not null" json:"last_value"`
	LastEventAt time.Time `gorm:"not null" json:"last_event_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
