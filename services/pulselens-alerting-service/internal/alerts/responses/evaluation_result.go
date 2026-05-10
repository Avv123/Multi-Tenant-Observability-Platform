package responses

type EvaluationResult struct {
	RuleID         string  `json:"rule_id"`
	Breached       bool    `json:"breached"`
	ObservedValue  float64 `json:"observed_value"`
	Comparator     string  `json:"comparator"`
	Threshold      float64 `json:"threshold"`
	IncidentStatus string  `json:"incident_status"`
}
