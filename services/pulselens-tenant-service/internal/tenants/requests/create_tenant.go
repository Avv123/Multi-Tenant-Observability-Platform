package requests

import "github.com/omniful/pulselens-tenant-service/internal/tenants/models"

type CreateTenantRequest struct {
	Name          string            `json:"name" validate:"required"`
	Slug          string            `json:"slug" validate:"required"`
	Plan          models.TenantPlan `json:"plan" validate:"required,oneof=free pro enterprise"`
	IngestQuota   int64             `json:"ingest_quota"`
	RetentionDays int               `json:"retention_days"`
	AdminName     string            `json:"admin_name" validate:"required"`
	AdminEmail    string            `json:"admin_email" validate:"required,email"`
	AdminPassword string            `json:"admin_password" validate:"required,min=8"`
}
