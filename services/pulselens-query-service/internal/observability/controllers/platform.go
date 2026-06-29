package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omniful/pulselens-platform/errs"
	platformresponse "github.com/omniful/pulselens-platform/response"
	observabilityresponses "github.com/omniful/pulselens-query-service/internal/observability/responses"
)

func (c *Controller) PlatformRuntime(ctx *gin.Context) {
	rows, err := c.service.PlatformRuntime(ctx)
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.Success(ctx, rows)
}


func (c *Controller) PlatformDependencies(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.DependencyHealth(ctx))
}

func (c *Controller) PlatformKafkaLag(ctx *gin.Context) {
	rows, err := c.service.KafkaLag(ctx)
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) PlatformOverview(ctx *gin.Context) {
	runtimeRows, runtimeErr := c.service.PlatformRuntime(ctx)
	if runtimeErr != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", runtimeErr.Error()))
		return
	}
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	cleanupRuns, cleanupErr := c.service.CleanupRuns(ctx, limit)
	if cleanupErr != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", cleanupErr.Error()))
		return
	}
	dependencies := c.service.DependencyHealth(ctx)
	lagRows, lagErr := c.service.KafkaLag(ctx)
	if lagErr != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", lagErr.Error()))
		return
	}
	platformresponse.Success(ctx, observabilityresponses.PlatformOverview{
		Runtime:      runtimeRows,
		CleanupRuns:  cleanupRuns,
		Dependencies: dependencies,
		KafkaLag:     lagRows,
	})
}

func (c *Controller) CleanupRuns(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	rows, err := c.service.CleanupRuns(ctx, limit)
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.Success(ctx, rows)
}
