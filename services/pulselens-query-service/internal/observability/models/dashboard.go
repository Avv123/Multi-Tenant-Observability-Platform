package models

import "time"

type Dashboard struct {
	ID               string    `gorm:"primaryKey;type:text" json:"id"`
	TenantID         string    `gorm:"index;not null" json:"tenant_id"`
	Name             string    `gorm:"not null" json:"name"`
	Description      string    `json:"description"`
	DefaultTimeRange string    `gorm:"not null;default:'120m'" json:"default_time_range"`
	Layout           string    `gorm:"type:jsonb;not null" json:"layout"`
	Widgets          string    `gorm:"type:jsonb;not null" json:"widgets"`
	CreatedBy        string    `gorm:"index" json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type DashboardWidget struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Metric    string         `json:"metric,omitempty"`
	Dataset   string         `json:"dataset,omitempty"`
	Title     string         `json:"title,omitempty"`
	Filters   map[string]any `json:"filters,omitempty"`
	Layout    map[string]any `json:"layout,omitempty"`
	ValueKey  string         `json:"value_key,omitempty"`
	LabelKey  string         `json:"label_key,omitempty"`
	ChartType string         `json:"chart_type,omitempty"`
}
