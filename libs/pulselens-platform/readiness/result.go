package readiness

type DependencyStatus struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Status    string `json:"status"`
	CheckedAt string `json:"checked_at"`
	Error     string `json:"error,omitempty"`
}

func Healthy(status DependencyStatus) bool {
	return status.Status == "healthy"
}
