package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	commonauth "github.com/omniful/pulselens-common/auth"
	pulsetenant "github.com/omniful/pulselens-common/tenant"
	"github.com/omniful/pulselens-platform/config"
	platformerrs "github.com/omniful/pulselens-platform/errs"
	platformresponse "github.com/omniful/pulselens-platform/response"
	"github.com/omniful/pulselens-tenant-service/internal/tenants/repositories"
	tenantrequests "github.com/omniful/pulselens-tenant-service/internal/tenants/requests"
	tenantservices "github.com/omniful/pulselens-tenant-service/internal/tenants/services"
	tenanterror "github.com/omniful/pulselens-tenant-service/pkg/error"
	"github.com/omniful/pulselens-tenant-service/pkg/postgres"
)

type Controller struct {
	service *tenantservices.Service
}

func NewController(_ context.Context) (*Controller, error) {
	repository := repositories.NewRepository(postgres.Get())
	service := tenantservices.New(
		repository,
		config.GetString("jwt.secret"),
		config.GetInt("jwt.expiryMinutes"),
	)
	return &Controller{service: service}, nil
}

func (c *Controller) CreateTenant(ctx *gin.Context) {
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

func (c *Controller) ListTenants(ctx *gin.Context) {
	rows, customError := c.service.ListTenants(ctx)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) GetTenant(ctx *gin.Context) {
	if strings.HasPrefix(ctx.FullPath(), "/admin/") && !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	row, customError := c.service.GetTenant(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) CreateService(ctx *gin.Context) {
	if strings.HasPrefix(ctx.FullPath(), "/admin/") && !authorizeTenant(ctx, ctx.Param("tenant_id")) {
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

func (c *Controller) ListServices(ctx *gin.Context) {
	if strings.HasPrefix(ctx.FullPath(), "/admin/") && !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	rows, customError := c.service.ListServices(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) CreateAPIKey(ctx *gin.Context) {
	var request tenantrequests.CreateAPIKeyRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	if strings.HasPrefix(ctx.FullPath(), "/admin/") && !authorizeTenant(ctx, request.TenantID) {
		return
	}
	row, customError := c.service.CreateAPIKey(ctx, actorUserID(ctx), &request)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

func (c *Controller) ListAPIKeys(ctx *gin.Context) {
	if strings.HasPrefix(ctx.FullPath(), "/admin/") && !authorizeTenant(ctx, ctx.Param("tenant_id")) {
		return
	}
	rows, customError := c.service.ListAPIKeys(ctx, ctx.Param("tenant_id"))
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) ResolveAPIKey(ctx *gin.Context) {
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

func (c *Controller) Login(ctx *gin.Context) {
	var request tenantrequests.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.BadRequest, err.Error()))
		return
	}
	row, customError := c.service.Login(ctx, &request)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) Me(ctx *gin.Context) {
	claims, ok := claimsFromContext(ctx)
	if !ok {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.Unauthorized, "missing auth claims"))
		return
	}
	row, customError := c.service.Me(ctx, claims)
	if customError.Exists() {
		tenanterror.NewErrorResponse(ctx, customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) ListAuditLogs(ctx *gin.Context) {
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

func (c *Controller) CreateUser(ctx *gin.Context) {
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

func (c *Controller) ListUsers(ctx *gin.Context) {
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

func claimsFromContext(ctx *gin.Context) (*commonauth.Claims, bool) {
	value, exists := ctx.Get("claims")
	if !exists {
		return nil, false
	}
	claims, ok := value.(*commonauth.Claims)
	return claims, ok
}

func actorUserID(ctx *gin.Context) string {
	claims, ok := claimsFromContext(ctx)
	if !ok {
		return ""
	}
	return claims.UserID
}

func authorizeTenant(ctx *gin.Context, tenantID string) bool {
	claims, ok := claimsFromContext(ctx)
	if !ok {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.Unauthorized, "missing auth claims"))
		return false
	}
	if claims.TenantID != tenantID {
		tenanterror.NewErrorResponse(ctx, platformerrs.New(tenanterror.Forbidden, "tenant access denied"))
		return false
	}
	return true
}
