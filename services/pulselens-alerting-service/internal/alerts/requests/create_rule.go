package requests

type CreateRuleRequest struct {
	ServiceID       string  `json:"service_id"`
	PolicyID        string  `json:"policy_id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	SignalType      string  `json:"signal_type"`
	MetricName      string  `json:"metric_name"`
	Severity        string  `json:"severity"`
	Aggregation     string  `json:"aggregation"`
	Comparator      string  `json:"comparator"`
	Threshold       float64 `json:"threshold"`
	WindowMinutes   int     `json:"window_minutes"`
	CooldownMinutes int     `json:"cooldown_minutes"`
	Active          *bool   `json:"active"`
}
