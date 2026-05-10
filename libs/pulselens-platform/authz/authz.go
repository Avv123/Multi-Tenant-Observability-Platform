package authz

import (
	"net/http"

	"github.com/gin-gonic/gin"
	commonauth "github.com/omniful/pulselens-common/auth"
	"github.com/omniful/pulselens-platform/errs"
	platformresponse "github.com/omniful/pulselens-platform/response"
)

const (
	RoleTenantAdmin  = "tenant_admin"
	RoleViewer       = "viewer"
	RoleOperator     = "operator"
	RoleAlertManager = "alert_manager"
	RoleServiceOwner = "service_owner"
)

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(ctx *gin.Context) {
		value, exists := ctx.Get("claims")
		if !exists {
			platformresponse.Error(ctx, http.StatusUnauthorized, errs.New("UNAUTHORIZED", "missing auth claims"))
			ctx.Abort()
			return
		}

		claims, ok := value.(*commonauth.Claims)
		if !ok || claims == nil {
			platformresponse.Error(ctx, http.StatusUnauthorized, errs.New("UNAUTHORIZED", "invalid auth claims"))
			ctx.Abort()
			return
		}

		if _, ok = allowed[claims.Role]; !ok {
			platformresponse.Error(ctx, http.StatusForbidden, errs.New("FORBIDDEN", "insufficient role"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
