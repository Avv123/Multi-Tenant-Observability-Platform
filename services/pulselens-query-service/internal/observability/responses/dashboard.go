package responses

import "time"

type Dashboard struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	DefaultTimeRange string            `json:"default_time_range"`
	Layout           interface{}       `json:"layout"`
	Widgets          []DashboardWidget `json:"widgets"`
	CreatedBy        string            `json:"created_by"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type DashboardWidget struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Type      string         `json:"type"`
	Dataset   string         `json:"dataset"`
	ChartType string         `json:"chart_type"`
	Metric    string         `json:"metric"`
	Filters   map[string]any `json:"filters,omitempty"`
	Layout    map[string]any `json:"layout,omitempty"`
	ValueKey  string         `json:"value_key,omitempty"`
	LabelKey  string         `json:"label_key,omitempty"`
}
