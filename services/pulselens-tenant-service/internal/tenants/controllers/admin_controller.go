package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omniful/pulselens-platform/config"
	platformerrs "github.com/omniful/pulselens-platform/errs"
	platformresponse "github.com/omniful/pulselens-platform/response"
	"github.com/omniful/pulselens-tenant-service/internal/tenants/repositories"
	tenantrequests "github.com/omniful/pulselens-tenant-service/internal/tenants/requests"
	tenantservices "github.com/omniful/pulselens-tenant-service/internal/tenants/services"
	tenanterror "github.com/omniful/pulselens-tenant-service/pkg/error"
	"github.com/omniful/pulselens-tenant-service/pkg/postgres"
)

// AdminController handles user-facing dashboard routes (/admin/api/v1).
// All routes here require a valid JWT AND the user's tenant must match the
// requested :tenant_id (enforced via authorizeTenant from helpers.go).
// This prevents cross-tenant data leakage (IDOR vulnerabilities).
type AdminController struct {
	service *tenantservices.Service
}

func NewAdminController(_ context.Context) (*AdminController, error) {
	repository := repositories.NewRepository(postgres.Get())
	service := tenantservices.New(
		repository,
		config.GetString("jwt.secret"),
		config.GetInt("jwt.expiryMinutes"),
	)
	return &AdminController{service: service}, nil
}

// GetTenant fetches a single tenant. Enforces that the caller belongs to that tenant.
func (c *AdminController) GetTenant(ctx *gin.Context) {
	if !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	row, customError := c.service.GetTenant(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, row)
}

// ListServices returns all services registered under a tenant.
func (c *AdminController) ListServices(ctx *gin.Context) {
	if !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	rows, customError := c.service.ListServices(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

// CreateService registers a new service under the caller's tenant.
func (c *AdminController) CreateService(ctx *gin.Context) {
	if !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	var request tenantrequests.CreateServiceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	row, customError := c.service.CreateService(ctx, ctx.Param("tenant_id"), actorUserID(ctx), &request)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

// ListAPIKeys returns all API keys for the caller's tenant.
func (c *AdminController) ListAPIKeys(ctx *gin.Context) {
	tenantID := ctx.Param("tenant_id")
	if tenantID != "" {
		if !authorizeTenant(ctx, tenantID) {
			return
		}
	} else {
		claims, ok := claimsFromContext(ctx)
		if !ok {
			tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.Unauthorized, "missing auth claims"))
			return
		}
		tenantID = claims.TenantID
	}
	rows, customError := c.service.ListAPIKeys(ctx, tenantID)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

// CreateAPIKey generates a new ingest API key. Enforces the caller owns the target tenant.
func (c *AdminController) CreateAPIKey(ctx *gin.Context) {
	var request tenantrequests.CreateAPIKeyRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	if !authorizeTenant(ctx, request.TenantID) {
		return
	}
	row, customError := c.service.CreateAPIKey(ctx, actorUserID(ctx), &request)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

// RotateAPIKey replaces a compromised API key with a new one, invalidating the old one.
// The new key is returned only once — it cannot be retrieved again.
func (c *AdminController) RotateAPIKey(ctx *gin.Context) {
	claims, ok := claimsFromContext(ctx)
	if !ok {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.Unauthorized, "missing auth claims"))
		return
	}
	var request tenantrequests.RotateAPIKeyRequest
	_ = ctx.ShouldBindJSON(&request)
	row, customError := c.service.RotateAPIKey(ctx, claims.UserID, claims.TenantID, ctx.Param("key_id"), request.Name)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, row)
}

// RevokeAPIKey permanently disables an API key, blocking all future ingest requests using it.
func (c *AdminController) RevokeAPIKey(ctx *gin.Context) {
	claims, ok := claimsFromContext(ctx)
	if !ok {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.Unauthorized, "missing auth claims"))
		return
	}
	customError := c.service.RevokeAPIKey(ctx, claims.UserID, claims.TenantID, ctx.Param("key_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, gin.H{"id": ctx.Param("key_id"), "status": "revoked"})
}

// ListAuditLogs returns the control-plane action history for a tenant.
func (c *AdminController) ListAuditLogs(ctx *gin.Context) {
	if !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	rows, customError := c.service.ListAuditLogs(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

// CreateUser adds a new human user account under the caller's tenant.
func (c *AdminController) CreateUser(ctx *gin.Context) {
	if !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	var request tenantrequests.CreateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	row, customError := c.service.CreateUser(ctx, ctx.Param("tenant_id"), actorUserID(ctx), &request)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

// ListUsers returns all human user accounts in the caller's tenant.
func (c *AdminController) ListUsers(ctx *gin.Context) {
	if !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	rows, customError := c.service.ListUsers(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}
