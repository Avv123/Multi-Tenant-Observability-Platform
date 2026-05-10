package responses

import pulsetenant "github.com/omniful/pulselens-common/tenant"

type CreateAPIKeyResponse struct {
	ID        string                    `json:"id"`
	TenantID  string                    `json:"tenant_id"`
	ServiceID string                    `json:"service_id"`
	Name      string                    `json:"name"`
	Key       string                    `json:"key"`
	KeyPrefix string                    `json:"key_prefix"`
	Scopes    []pulsetenant.APIKeyScope `json:"scopes"`
}
