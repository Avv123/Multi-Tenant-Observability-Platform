package router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/omniful/pulselens-platform/cors"
	"github.com/omniful/pulselens-platform/httpserver"
	"github.com/omniful/pulselens-platform/middleware"
)

func Initialize(_ context.Context, server *httpserver.Server) error {
	server.Engine.Use(cors.Middleware())
	server.Engine.Use(middleware.RequestID())
	server.Engine.Use(gin.Recovery())
	server.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return nil
}
