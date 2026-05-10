package requests

import pulsetenant "github.com/omniful/pulselens-common/tenant"

type CreateAPIKeyRequest struct {
	TenantID  string                    `json:"tenant_id" validate:"required"`
	ServiceID string                    `json:"service_id" validate:"required"`
	Name      string                    `json:"name" validate:"required"`
	Scopes    []pulsetenant.APIKeyScope `json:"scopes" validate:"required,min=1"`
}
