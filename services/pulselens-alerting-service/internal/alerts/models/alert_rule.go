package models

import "time"

type AlertRule struct {
	ID               string     `gorm:"primaryKey" json:"id"`
	TenantID         string     `gorm:"index;not null" json:"tenant_id"`
	ServiceID        string     `gorm:"index" json:"service_id"`
	PolicyID         string     `gorm:"index" json:"policy_id"`
	Name             string     `gorm:"not null" json:"name"`
	Description      string     `json:"description"`
	SignalType       string     `gorm:"index;not null" json:"signal_type"`
	MetricName       string     `gorm:"index" json:"metric_name"`
	Severity         string     `gorm:"index" json:"severity"`
	Aggregation      string     `json:"aggregation"`
	Comparator       string     `json:"comparator"`
	Threshold        float64    `json:"threshold"`
	WindowMinutes    int        `json:"window_minutes"`
	CooldownMinutes  int        `json:"cooldown_minutes"`
	Active           bool       `gorm:"default:true" json:"active"`
	LastTriggeredAt  *time.Time `json:"last_triggered_at"`
	LastEvaluationAt *time.Time `json:"last_evaluation_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
