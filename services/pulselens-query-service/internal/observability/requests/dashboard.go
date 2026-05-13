package requests

type CreateDashboardRequest struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	DefaultTimeRange string            `json:"default_time_range"`
	Layout           interface{}       `json:"layout"`
	Widgets          []DashboardWidget `json:"widgets"`
}

type UpdateDashboardRequest struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	DefaultTimeRange string            `json:"default_time_range"`
	Layout           interface{}       `json:"layout"`
	Widgets          []DashboardWidget `json:"widgets"`
}

type UpdateDashboardWidgetRequest struct {
	Title     string         `json:"title"`
	Type      string         `json:"type"`
	Dataset   string         `json:"dataset"`
	ChartType string         `json:"chart_type"`
	Metric    string         `json:"metric"`
	Filters   map[string]any `json:"filters"`
	Layout    map[string]any `json:"layout"`
}

type DashboardWidget struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Type      string         `json:"type"`
	Dataset   string         `json:"dataset"`
	ChartType string         `json:"chart_type"`
	Metric    string         `json:"metric"`
	Filters   map[string]any `json:"filters"`
	Layout    map[string]any `json:"layout"`
}
