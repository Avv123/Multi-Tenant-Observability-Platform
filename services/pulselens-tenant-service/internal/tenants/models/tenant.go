package models

import "time"

type TenantPlan string

const (
	PlanFree       TenantPlan = "free"
	PlanPro        TenantPlan = "pro"
	PlanEnterprise TenantPlan = "enterprise"
)

type Tenant struct {
	ID            string     `gorm:"primaryKey;type:text" json:"id"`
	Name          string     `gorm:"not null" json:"name"`
	Slug          string     `gorm:"uniqueIndex;not null" json:"slug"`
	Plan          TenantPlan `gorm:"not null" json:"plan"`
	IngestQuota   int64      `gorm:"not null" json:"ingest_quota"`
	RetentionDays int        `gorm:"not null" json:"retention_days"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
