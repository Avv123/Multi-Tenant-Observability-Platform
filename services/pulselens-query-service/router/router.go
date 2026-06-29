package router

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/omniful/pulselens-platform/authz"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/cors"
	"github.com/omniful/pulselens-platform/httpserver"
	"github.com/omniful/pulselens-platform/middleware"
	platformreadiness "github.com/omniful/pulselens-platform/readiness"
	appinit "github.com/omniful/pulselens-query-service/init"
	querycontrollers "github.com/omniful/pulselens-query-service/internal/observability/controllers"
	querymiddleware "github.com/omniful/pulselens-query-service/internal/observability/middleware"
)

func Initialize(ctx context.Context, server *httpserver.Server) error {
	server.Engine.Use(cors.Middleware())
	server.Engine.Use(middleware.RequestID())
	server.Engine.Use(gin.Recovery())
	server.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	server.Engine.GET("/ready", func(c *gin.Context) {
		rows := appinit.Readiness(c.Request.Context())
		status := http.StatusOK
		if !allHealthy(rows) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"status": readinessStatus(rows), "dependencies": rows})
	})

	controller, err := querycontrollers.NewController(ctx)
	if err != nil {
		return err
	}

	api := server.Engine.Group("/api/v1")
	api.Use(querymiddleware.AuthenticateJWT(ctx))
	api.GET("/overview", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.Overview)
	api.GET("/services/health", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.ServiceHealth)
	api.GET("/analytics/log-severity", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.LogSeveritySeries)
	api.GET("/analytics/metric-series", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.MetricSeries)
	api.GET("/analytics/trace-latency", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.TraceLatencySeries)
	api.GET("/transactions", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.ListTransactions)
	api.GET("/errors/groups", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.ListErrorGroups)
	api.GET("/service-map", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.GetServiceMap)
	api.GET("/logs", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.ListLogs)
	api.GET("/metrics", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.ListMetrics)
	api.GET("/traces", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.ListTraces)
	api.GET("/traces/:trace_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.TraceDetail)
	api.GET("/saved-queries", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.ListSavedQueries)
	api.POST("/saved-queries", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner), controller.CreateSavedQuery)
	api.PATCH("/saved-queries/:query_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner), controller.UpdateSavedQuery)
	api.GET("/dashboards", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleViewer, authz.RoleOperator, authz.RoleAlertManager, authz.RoleServiceOwner), controller.ListDashboards)
	api.POST("/dashboards", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner), controller.CreateDashboard)
	api.PUT("/dashboards/:dashboard_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner), controller.UpdateDashboard)
	api.PATCH("/dashboards/:dashboard_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner), controller.UpdateDashboard)
	api.PATCH("/dashboards/:dashboard_id/widgets/:widget_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner), controller.UpdateDashboardWidget)
	api.DELETE("/dashboards/:dashboard_id/widgets/:widget_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner), controller.DeleteDashboardWidget)
	api.GET("/platform/runtime", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleAlertManager), controller.PlatformRuntime)
	api.GET("/platform/dependencies", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleAlertManager), controller.PlatformDependencies)
	api.GET("/platform/kafka-lag", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleAlertManager), controller.PlatformKafkaLag)
	api.GET("/platform/overview", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleAlertManager), controller.PlatformOverview)
	api.GET("/platform/cleanup-runs", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleAlertManager), controller.CleanupRuns)

	uiDir, err := filepath.Abs(filepath.Clean(config.GetString("ui.dir")))
	if err != nil {
		return err
	}
	server.Engine.Static("/assets", filepath.Join(uiDir, "assets"))
	server.Engine.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(uiDir, "index.html"))
	})
	server.Engine.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet && c.GetHeader("Accept") != "application/json" {
			c.File(filepath.Join(uiDir, "index.html"))
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	})
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
