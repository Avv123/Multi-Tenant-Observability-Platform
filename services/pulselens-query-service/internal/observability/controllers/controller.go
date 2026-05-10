package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	commonauth "github.com/omniful/pulselens-common/auth"
	"github.com/omniful/pulselens-platform/errs"
	platformresponse "github.com/omniful/pulselens-platform/response"
	observabilityrequests "github.com/omniful/pulselens-query-service/internal/observability/requests"
	observabilityservices "github.com/omniful/pulselens-query-service/internal/observability/services"
)

type Controller struct {
	service *observabilityservices.Service
}

func NewController(_ context.Context) (*Controller, error) {
	return &Controller{service: observabilityservices.New()}, nil
}

func (c *Controller) Overview(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.Overview(ctx, claimsFromContext(ctx)))
}

func (c *Controller) ListLogs(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.ListLogs(ctx, claimsFromContext(ctx), buildFilters(ctx)))
}

func (c *Controller) ListMetrics(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.ListMetrics(ctx, claimsFromContext(ctx), buildFilters(ctx)))
}

func (c *Controller) ListTraces(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.ListTraces(ctx, claimsFromContext(ctx), buildFilters(ctx)))
}

func (c *Controller) ServiceHealth(ctx *gin.Context) {
	filters := buildFilters(ctx)
	platformresponse.Success(ctx, c.service.ServiceHealth(ctx, claimsFromContext(ctx), filters.Limit))
}

func (c *Controller) LogSeveritySeries(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.LogSeveritySeries(ctx, claimsFromContext(ctx), buildFilters(ctx)))
}

func (c *Controller) MetricSeries(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.MetricSeries(ctx, claimsFromContext(ctx), buildFilters(ctx)))
}

func (c *Controller) TraceLatencySeries(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.TraceLatencySeries(ctx, claimsFromContext(ctx), buildFilters(ctx)))
}

func (c *Controller) TraceDetail(ctx *gin.Context) {
	platformresponse.Success(ctx, c.service.TraceDetail(ctx, claimsFromContext(ctx), ctx.Param("trace_id")))
}

func (c *Controller) CreateSavedQuery(ctx *gin.Context) {
	var request observabilityrequests.CreateSavedQueryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, errs.New("BAD_REQUEST", err.Error()))
		return
	}
	row, err := c.service.CreateSavedQuery(ctx, claimsFromContext(ctx), request)
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

func (c *Controller) ListSavedQueries(ctx *gin.Context) {
	rows, err := c.service.ListSavedQueries(ctx, claimsFromContext(ctx))
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) UpdateSavedQuery(ctx *gin.Context) {
	var request observabilityrequests.UpdateSavedQueryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, errs.New("BAD_REQUEST", err.Error()))
		return
	}
	row, err := c.service.UpdateSavedQuery(ctx, claimsFromContext(ctx), ctx.Param("query_id"), request)
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) CreateDashboard(ctx *gin.Context) {
	var request observabilityrequests.CreateDashboardRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, errs.New("BAD_REQUEST", err.Error()))
		return
	}
	row, err := c.service.CreateDashboard(ctx, claimsFromContext(ctx), request)
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.WithStatus(ctx, http.StatusCreated, row, nil)
}

func (c *Controller) ListDashboards(ctx *gin.Context) {
	rows, err := c.service.ListDashboards(ctx, claimsFromContext(ctx))
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) UpdateDashboard(ctx *gin.Context) {
	var request observabilityrequests.UpdateDashboardRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, errs.New("BAD_REQUEST", err.Error()))
		return
	}
	row, err := c.service.UpdateDashboard(ctx, claimsFromContext(ctx), ctx.Param("dashboard_id"), request)
	if err != nil {
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.Success(ctx, row)
}

func claimsFromContext(ctx *gin.Context) *commonauth.Claims {
	value, _ := ctx.Get("claims")
	if claims, ok := value.(*commonauth.Claims); ok {
		return claims
	}
	return &commonauth.Claims{}
}

func buildFilters(ctx *gin.Context) observabilityrequests.Filters {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	startTime, _ := time.Parse(time.RFC3339, ctx.Query("start_time"))
	endTime, _ := time.Parse(time.RFC3339, ctx.Query("end_time"))
	return observabilityrequests.Filters{
		ServiceID:   ctx.Query("service_id"),
		Environment: ctx.Query("environment"),
		Severity:    ctx.Query("severity"),
		MetricName:  ctx.Query("metric_name"),
		Search:      ctx.Query("search"),
		TraceID:     ctx.Query("trace_id"),
		Limit:       limit,
		Offset:      offset,
		StartTime:   startTime,
		EndTime:     endTime,
	}
}
