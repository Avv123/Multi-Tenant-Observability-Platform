package router

import (
	"context"

	"github.com/gin-gonic/gin"
	alertcontrollers "github.com/omniful/pulselens-alerting-service/internal/alerts/controllers"
	alertmiddleware "github.com/omniful/pulselens-alerting-service/internal/alerts/middleware"
	"github.com/omniful/pulselens-platform/authz"
	"github.com/omniful/pulselens-platform/cors"
	"github.com/omniful/pulselens-platform/httpserver"
	platformmiddleware "github.com/omniful/pulselens-platform/middleware"
)

func Initialize(ctx context.Context, server *httpserver.Server) error {
	server.Engine.Use(cors.Middleware())
	server.Engine.Use(platformmiddleware.RequestID())
	server.Engine.Use(gin.Recovery())
	server.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	controller, err := alertcontrollers.NewController(ctx)
	if err != nil {
		return err
	}

	api := server.Engine.Group("/api/v1")
	api.Use(alertmiddleware.AuthenticateJWT(ctx))
	api.GET("/alert-rules", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator, authz.RoleViewer, authz.RoleServiceOwner), controller.ListRules)
	api.POST("/alert-rules", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager), controller.CreateRule)
	api.PATCH("/alert-rules/:rule_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager), controller.UpdateRule)
	api.GET("/alert-policies", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator, authz.RoleViewer), controller.ListPolicies)
	api.POST("/alert-policies", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager), controller.CreatePolicy)
	api.PATCH("/alert-policies/:policy_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager), controller.UpdatePolicy)
	api.GET("/incidents", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator, authz.RoleViewer, authz.RoleServiceOwner), controller.ListIncidents)
	api.GET("/incidents/:incident_id", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator, authz.RoleViewer, authz.RoleServiceOwner), controller.GetIncident)
	api.GET("/incidents/:incident_id/timeline", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator, authz.RoleViewer, authz.RoleServiceOwner), controller.ListIncidentTimeline)
	api.GET("/incidents/:incident_id/deliveries", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator, authz.RoleViewer, authz.RoleServiceOwner), controller.ListIncidentDeliveries)
	api.POST("/incidents/:incident_id/assign", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager), controller.AssignIncident)
	api.POST("/incidents/:incident_id/acknowledge", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator), controller.AcknowledgeIncident)
	api.POST("/incidents/:incident_id/resolve", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator), controller.ResolveIncident)
	api.GET("/incidents/:incident_id/comments", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator, authz.RoleViewer, authz.RoleServiceOwner), controller.ListIncidentComments)
	api.POST("/incidents/:incident_id/comments", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator), controller.AddIncidentComment)
	api.GET("/notification-channels", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator, authz.RoleViewer), controller.ListNotificationChannels)
	api.POST("/notification-channels", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager), controller.CreateNotificationChannel)
	api.GET("/notification-deliveries", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleAlertManager, authz.RoleOperator), controller.ListNotificationDeliveries)
	return nil
}
