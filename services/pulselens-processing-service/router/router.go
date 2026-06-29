package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Avv123/pulselens-platform/cors"
	"github.com/Avv123/pulselens-platform/httpserver"
	"github.com/Avv123/pulselens-platform/middleware"
	platformreadiness "github.com/Avv123/pulselens-platform/readiness"
	appinit "github.com/Avv123/pulselens-processing-service/init"
)

func Initialize(_ context.Context, server *httpserver.Server) error {
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
