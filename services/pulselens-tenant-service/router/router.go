package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Avv123/pulselens-platform/authz"
	"github.com/Avv123/pulselens-platform/cors"
	"github.com/Avv123/pulselens-platform/httpserver"
	"github.com/Avv123/pulselens-platform/middleware"
	platformreadiness "github.com/Avv123/pulselens-platform/readiness"
	appinit "github.com/Avv123/pulselens-tenant-service/init"
	tenantcontrollers "github.com/Avv123/pulselens-tenant-service/internal/tenants/controllers"
)

func Initialize(ctx context.Context, s *httpserver.Server) error {
	s.Engine.Use(cors.Middleware())
	s.Engine.Use(middleware.RequestID())
	s.Engine.Use(gin.Recovery())
	s.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	s.Engine.GET("/ready", func(c *gin.Context) {
		rows := appinit.Readiness(c.Request.Context())
		status := http.StatusOK
		if !allHealthy(rows) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"status": readinessStatus(rows), "dependencies": rows})
	})

	authCtrl, err := tenantcontrollers.NewAuthController(ctx)
	if err != nil {
		return err
	}

	internalCtrl, err := tenantcontrollers.NewInternalController(ctx)
	if err != nil {
		return err
	}

	adminCtrl, err := tenantcontrollers.NewAdminController(ctx)
	if err != nil {
		return err
	}

	// Public auth routes — no token required for login, JWT required for /me
	api := s.Engine.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authCtrl.Login)
			auth.GET("/me", tenantcontrollers.AuthenticateJWT(ctx), authCtrl.Me)
		}
	}

	// Internal service-to-service routes — protected by static INTERNAL_TOKEN only.
	// Called by backend services (e.g. ingest-service). No user JWT or tenant scoping needed.
	internal := s.Engine.Group("/internal/api/v1")
	internal.Use(tenantcontrollers.AuthenticateInternalToken(ctx))
	{
		internal.POST("/tenants", internalCtrl.CreateTenant)
		internal.GET("/tenants", internalCtrl.ListTenants)
		internal.GET("/tenants/:tenant_id", internalCtrl.GetTenant)
		internal.POST("/tenants/:tenant_id/services", internalCtrl.CreateService)
		internal.GET("/tenants/:tenant_id/services", internalCtrl.ListServices)
		internal.POST("/api-keys", internalCtrl.CreateAPIKey)
		internal.GET("/tenants/:tenant_id/api-keys", internalCtrl.ListAPIKeys)
		internal.POST("/auth/resolve-api-key", internalCtrl.ResolveAPIKey)
	}

	// Admin dashboard routes — protected by JWT and tenant-admin role.
	// Every handler enforces that the caller's tenant matches the requested :tenant_id.
	admin := s.Engine.Group("/admin/api/v1")
	admin.Use(tenantcontrollers.AuthenticateJWT(ctx))
	admin.Use(authz.RequireRoles(authz.RoleTenantAdmin))
	{
		admin.GET("/api-keys", adminCtrl.ListAPIKeys)
		admin.GET("/tenants/:tenant_id", adminCtrl.GetTenant)
		admin.GET("/tenants/:tenant_id/services", adminCtrl.ListServices)
		admin.GET("/tenants/:tenant_id/api-keys", adminCtrl.ListAPIKeys)
		admin.GET("/tenants/:tenant_id/audit-logs", adminCtrl.ListAuditLogs)
		admin.GET("/tenants/:tenant_id/users", adminCtrl.ListUsers)
		admin.POST("/tenants/:tenant_id/services", adminCtrl.CreateService)
		admin.POST("/tenants/:tenant_id/users", adminCtrl.CreateUser)
		admin.POST("/api-keys", adminCtrl.CreateAPIKey)
		admin.POST("/api-keys/:key_id/rotate", adminCtrl.RotateAPIKey)
		admin.POST("/api-keys/:key_id/revoke", adminCtrl.RevokeAPIKey)
	}

	return nil
}

func allHealthy(rows []platformreadiness.DependencyStatus) bool {
	for _, row := range rows {
		if row.Status != "healthy" {
			return false
		}
	}
	return true
}

func readinessStatus(rows []platformreadiness.DependencyStatus) string {
	if allHealthy(rows) {
		return "ready"
	}
	return "degraded"
}
