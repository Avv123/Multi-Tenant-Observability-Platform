package router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/omniful/pulselens-archive-service/internal/replay/controllers"
	replaymiddleware "github.com/omniful/pulselens-archive-service/internal/replay/middleware"
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

	controller, err := controllers.NewController(ctx)
	if err != nil {
		return err
	}

	api := server.Engine.Group("/api/v1")
	api.Use(replaymiddleware.AuthenticateJWT(ctx))
	api.GET("/archive/stats", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner, authz.RoleAlertManager), controller.Stats)
	api.GET("/replay-jobs", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner, authz.RoleAlertManager), controller.ListReplayJobs)
	api.POST("/replay-jobs", authz.RequireRoles(authz.RoleTenantAdmin, authz.RoleOperator, authz.RoleServiceOwner), controller.CreateReplayJob)
	return nil
}
