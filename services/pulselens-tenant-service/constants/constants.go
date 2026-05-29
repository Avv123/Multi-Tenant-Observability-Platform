package constants

const (
	HeaderInternalToken = "X-Internal-Token"
	HeaderAPIKey        = "X-API-Key"
	ContextUser         = "user"
)

// Audit Log Actions
const (
	ActionTenantCreated  = "tenant.created"
	ActionTenantDeleted  = "tenant.deleted"
	ActionTenantUpdated  = "tenant.updated"
	ActionServiceCreated = "service.created"
	ActionServiceDeleted = "service.deleted"
	ActionAPIKeyCreated  = "api_key.created"
)

// Resource Types
const (
	ResourceTenant  = "tenant"
	ResourceService = "service"
	ResourceAPIKey  = "api_key"
)
