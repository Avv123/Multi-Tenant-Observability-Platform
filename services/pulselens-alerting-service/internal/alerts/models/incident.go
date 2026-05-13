package models

import "time"

type Incident struct {
	ID               string     `gorm:"primaryKey" json:"id"`
	AlertRuleID      string     `gorm:"index;not null" json:"alert_rule_id"`
	TenantID         string     `gorm:"index;not null" json:"tenant_id"`
	ServiceID        string     `gorm:"index" json:"service_id"`
	Severity         string     `gorm:"index" json:"severity"`
	Status           string     `gorm:"index;not null" json:"status"`
	AssignedTo       string     `gorm:"index" json:"assigned_to"`
	AssignedAt       *time.Time `json:"assigned_at"`
	AcknowledgedBy   string     `gorm:"index" json:"acknowledged_by"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at"`
	EscalationLevel  int        `gorm:"not null;default:0" json:"escalation_level"`
	EscalationCount  int        `gorm:"not null;default:0" json:"escalation_count"`
	LastEscalatedAt  *time.Time `json:"last_escalated_at"`
	NextEscalationAt *time.Time `json:"next_escalation_at"`
	Title            string     `json:"title"`
	Summary          string     `json:"summary"`
	ObservedValue    float64    `json:"observed_value"`
	Threshold        float64    `json:"threshold"`
	TriggeredAt      time.Time  `json:"triggered_at"`
	ResolvedAt       *time.Time `json:"resolved_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
