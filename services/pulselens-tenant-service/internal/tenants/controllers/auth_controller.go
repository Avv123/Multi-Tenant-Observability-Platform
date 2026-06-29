package controllers

import (
	"context"

	"github.com/gin-gonic/gin"
	platformerrs "github.com/Avv123/pulselens-platform/errs"
	platformresponse "github.com/Avv123/pulselens-platform/response"
	"github.com/Avv123/pulselens-platform/config"
	"github.com/Avv123/pulselens-tenant-service/internal/tenants/repositories"
	tenantservices "github.com/Avv123/pulselens-tenant-service/internal/tenants/services"
	tenantrequests "github.com/Avv123/pulselens-tenant-service/internal/tenants/requests"
	tenanterror "github.com/Avv123/pulselens-tenant-service/pkg/error"
	"github.com/Avv123/pulselens-tenant-service/pkg/postgres"
)

// AuthController handles public-facing authentication routes (/api/v1/auth).
// These routes are unauthenticated or require only a valid JWT (no tenant scoping).
type AuthController struct {
	service *tenantservices.Service
}

func NewAuthController(_ context.Context) (*AuthController, error) {
	repository := repositories.NewRepository(postgres.Get())
	service := tenantservices.New(
		repository,
		config.GetString("jwt.secret"),
		config.GetInt("jwt.expiryMinutes"),
	)
	return &AuthController{service: service}, nil
}

// Login accepts email/password and returns a signed JWT token.
func (c *AuthController) Login(ctx *gin.Context) {
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

// Me returns the profile and permissions of the currently authenticated user.
// Used by the React frontend on boot to determine who is logged in and what UI to show.
func (c *AuthController) Me(ctx *gin.Context) {
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
