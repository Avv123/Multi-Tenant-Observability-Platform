package requests

type CreateDashboardRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Layout      interface{} `json:"layout"`
	Widgets     interface{} `json:"widgets"`
}

type UpdateDashboardRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Layout      interface{} `json:"layout"`
	Widgets     interface{} `json:"widgets"`
}
