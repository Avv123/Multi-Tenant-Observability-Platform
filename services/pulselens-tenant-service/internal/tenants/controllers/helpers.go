package controllers

import (
	"github.com/gin-gonic/gin"
	commonauth "github.com/omniful/pulselens-common/auth"
	platformerrs "github.com/omniful/pulselens-platform/errs"
	tenanterror "github.com/omniful/pulselens-tenant-service/pkg/error"
)

// claimsFromContext extracts the JWT claims injected by the AuthenticateJWT middleware.
func claimsFromContext(ctx *gin.Context) (*commonauth.Claims, bool) {
	value, exists := ctx.Get("claims")
	if !exists {
		return nil, false
	}
	claims, ok := value.(*commonauth.Claims)
	return claims, ok
}

// actorUserID returns the user ID of the authenticated user from the JWT claims.
func actorUserID(ctx *gin.Context) string {
	claims, ok := claimsFromContext(ctx)
	if !ok {
		return ""
	}
	return claims.UserID
}

// authorizeTenant ensures the authenticated user belongs to the requested tenant.
// Prevents cross-tenant data access (IDOR) in admin routes.
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
