package requests

type CreateSavedQueryRequest struct {
	Name       string      `json:"name"`
	QueryType  string      `json:"query_type"`
	Definition interface{} `json:"definition"`
}

type UpdateSavedQueryRequest struct {
	Name       string      `json:"name"`
	QueryType  string      `json:"query_type"`
	Definition interface{} `json:"definition"`
}
