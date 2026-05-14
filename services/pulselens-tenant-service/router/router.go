package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omniful/pulselens-platform/authz"
	"github.com/omniful/pulselens-platform/cors"
	"github.com/omniful/pulselens-platform/httpserver"
	"github.com/omniful/pulselens-platform/middleware"
	platformreadiness "github.com/omniful/pulselens-platform/readiness"
	appinit "github.com/omniful/pulselens-tenant-service/init"
	tenantcontrollers "github.com/omniful/pulselens-tenant-service/internal/tenants/controllers"
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

	ctrl, err := tenantcontrollers.NewController(ctx)
	if err != nil {
		return err
	}

	api := s.Engine.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", ctrl.Login)
			auth.GET("/me", tenantcontrollers.AuthenticateJWT(ctx), ctrl.Me)
		}
	}

	internal := s.Engine.Group("/internal/api/v1")
	internal.Use(tenantcontrollers.AuthenticateInternalToken(ctx))
	{
		internal.POST("/tenants", ctrl.CreateTenant)
		internal.GET("/tenants", ctrl.ListTenants)
		internal.GET("/tenants/:tenant_id", ctrl.GetTenant)
		internal.POST("/tenants/:tenant_id/services", ctrl.CreateService)
		internal.GET("/tenants/:tenant_id/services", ctrl.ListServices)
		internal.POST("/api-keys", ctrl.CreateAPIKey)
		internal.GET("/tenants/:tenant_id/api-keys", ctrl.ListAPIKeys)
		internal.POST("/auth/resolve-api-key", ctrl.ResolveAPIKey)
	}

	admin := s.Engine.Group("/admin/api/v1")
	admin.Use(tenantcontrollers.AuthenticateJWT(ctx))
	admin.Use(authz.RequireRoles(authz.RoleTenantAdmin))
	{
		admin.GET("/api-keys", ctrl.ListAPIKeys)
		admin.GET("/tenants/:tenant_id", ctrl.GetTenant)
		admin.GET("/tenants/:tenant_id/services", ctrl.ListServices)
		admin.GET("/tenants/:tenant_id/api-keys", ctrl.ListAPIKeys)
		admin.GET("/tenants/:tenant_id/audit-logs", ctrl.ListAuditLogs)
		admin.GET("/tenants/:tenant_id/users", ctrl.ListUsers)
		admin.POST("/tenants/:tenant_id/services", ctrl.CreateService)
		admin.POST("/tenants/:tenant_id/users", ctrl.CreateUser)
		admin.POST("/api-keys", ctrl.CreateAPIKey)
		admin.POST("/api-keys/:key_id/rotate", ctrl.RotateAPIKey)
		admin.POST("/api-keys/:key_id/revoke", ctrl.RevokeAPIKey)
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
