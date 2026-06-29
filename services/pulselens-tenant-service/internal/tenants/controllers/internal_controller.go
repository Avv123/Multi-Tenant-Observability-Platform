package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	pulsetenant "github.com/Avv123/pulselens-common/tenant"
	"github.com/Avv123/pulselens-platform/config"
	platformerrs "github.com/Avv123/pulselens-platform/errs"
	platformresponse "github.com/Avv123/pulselens-platform/response"
	"github.com/Avv123/pulselens-tenant-service/internal/tenants/repositories"
	tenantrequests "github.com/Avv123/pulselens-tenant-service/internal/tenants/requests"
	tenantservices "github.com/Avv123/pulselens-tenant-service/internal/tenants/services"
	tenanterror "github.com/Avv123/pulselens-tenant-service/pkg/error"
	"github.com/Avv123/pulselens-tenant-service/pkg/postgres"
)

// InternalController handles service-to-service routes (/internal/api/v1).
// These routes are protected by a static INTERNAL_TOKEN and are only called
// by other backend services (e.g. ingest-service resolving an API key).
// No user JWT or tenant scoping is needed here — all backend services are fully trusted.
type InternalController struct {
	service *tenantservices.Service
}

func NewInternalController(_ context.Context) (*InternalController, error) {
	repository := repositories.NewRepository(postgres.Get())
	service := tenantservices.New(
		repository,
		config.GetString("jwt.secret"),
		config.GetInt("jwt.expiryMinutes"),
	)
	return &InternalController{service: service}, nil
}

// CreateTenant provisions a new isolated tenant and its first admin user.
func (c *InternalController) CreateTenant(ctx *gin.Context) {
	var request tenantrequests.CreateTenantRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	tenantModel, userModel, customError := c.service.CreateTenant(ctx, &request)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, gin.H{
		"tenant": tenantModel,
		"admin": gin.H{
			"id":    userModel.ID,
			"name":  userModel.Name,
			"email": userModel.Email,
			"role":  userModel.Role,
		},
	}, nil)
}

// ListTenants returns all tenants. Used by platform-level admin tooling.
func (c *InternalController) ListTenants(ctx *gin.Context) {
	rows, customError := c.service.ListTenants(ctx)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

// GetTenant fetches a single tenant by ID. No user scoping needed — trusted callers only.
func (c *InternalController) GetTenant(ctx *gin.Context) {
	row, customError := c.service.GetTenant(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, row)
}

// CreateService registers a new service under a tenant. Called by bootstrap scripts.
func (c *InternalController) CreateService(ctx *gin.Context) {
	var request tenantrequests.CreateServiceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	row, customError := c.service.CreateService(ctx, ctx.Param("tenant_id"), "", &request)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

// ListServices returns all services for a tenant. Used by ingest-service during resolution.
func (c *InternalController) ListServices(ctx *gin.Context) {
	rows, customError := c.service.ListServices(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

// CreateAPIKey generates a new ingest API key for a service. Called by bootstrap scripts.
func (c *InternalController) CreateAPIKey(ctx *gin.Context) {
	var request tenantrequests.CreateAPIKeyRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	row, customError := c.service.CreateAPIKey(ctx, "", &request)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

// ListAPIKeys returns all API keys for a given tenant.
func (c *InternalController) ListAPIKeys(ctx *gin.Context) {
	rows, customError := c.service.ListAPIKeys(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

// ResolveAPIKey is called by ingest-service to convert a raw API key string
// into a validated TenantID and ServiceID before processing any telemetry payload.
func (c *InternalController) ResolveAPIKey(ctx *gin.Context) {
	var request pulsetenant.ResolveAPIKeyRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	row, customError := c.service.ResolveAPIKey(ctx, request.APIKey)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, row)
}
