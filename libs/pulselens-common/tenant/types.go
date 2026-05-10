package tenant

type APIKeyScope string

const (
	ScopeIngest APIKeyScope = "ingest"
	ScopeQuery  APIKeyScope = "query"
	ScopeAdmin  APIKeyScope = "admin"
)

type ResolvedAPIKey struct {
	KeyID       string        `json:"key_id"`
	TenantID    string        `json:"tenant_id"`
	TenantName  string        `json:"tenant_name"`
	Plan        string        `json:"plan"`
	IngestQuota int64         `json:"ingest_quota"`
	ServiceID   string        `json:"service_id"`
	ServiceName string        `json:"service_name"`
	Environment string        `json:"environment"`
	Scopes      []APIKeyScope `json:"scopes"`
	Active      bool          `json:"active"`
}

type ResolveAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}
