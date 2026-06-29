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
	row, customError := c.service.Overview(ctx, claimsFromContext(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) ListLogs(ctx *gin.Context) {
	rows, customError := c.service.ListLogs(ctx, claimsFromContext(ctx), buildFilters(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) ListMetrics(ctx *gin.Context) {
	rows, customError := c.service.ListMetrics(ctx, claimsFromContext(ctx), buildFilters(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) ListTraces(ctx *gin.Context) {
	rows, customError := c.service.ListTraces(ctx, claimsFromContext(ctx), buildFilters(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) ServiceHealth(ctx *gin.Context) {
	filters := buildFilters(ctx)
	rows, customError := c.service.ServiceHealth(ctx, claimsFromContext(ctx), filters.Limit)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) LogSeveritySeries(ctx *gin.Context) {
	rows, customError := c.service.LogSeveritySeries(ctx, claimsFromContext(ctx), buildFilters(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) MetricSeries(ctx *gin.Context) {
	rows, customError := c.service.MetricSeries(ctx, claimsFromContext(ctx), buildFilters(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) TraceLatencySeries(ctx *gin.Context) {
	rows, customError := c.service.TraceLatencySeries(ctx, claimsFromContext(ctx), buildFilters(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) TraceDetail(ctx *gin.Context) {
	rows, customError := c.service.TraceDetail(ctx, claimsFromContext(ctx), ctx.Param("trace_id"))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
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

func (c *Controller) UpdateDashboardWidget(ctx *gin.Context) {
	var request observabilityrequests.UpdateDashboardWidgetRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		platformresponse.Error(ctx, http.StatusBadRequest, errs.New("BAD_REQUEST", err.Error()))
		return
	}
	row, err := c.service.UpdateDashboardWidget(ctx, claimsFromContext(ctx), ctx.Param("dashboard_id"), ctx.Param("widget_id"), request)
	if err != nil {
		if customError, ok := err.(errs.CustomError); ok {
			platformresponse.Error(ctx, statusCode(customError.Code()), customError)
			return
		}
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.Success(ctx, row)
}

func (c *Controller) DeleteDashboardWidget(ctx *gin.Context) {
	if err := c.service.DeleteDashboardWidget(ctx, claimsFromContext(ctx), ctx.Param("dashboard_id"), ctx.Param("widget_id")); err != nil {
		if customError, ok := err.(errs.CustomError); ok {
			platformresponse.Error(ctx, statusCode(customError.Code()), customError)
			return
		}
		platformresponse.Error(ctx, http.StatusInternalServerError, errs.New("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	platformresponse.WithStatus(ctx, http.StatusOK, gin.H{"deleted": true}, nil)
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

	// Accept explicit RFC3339 start/end times
	startTime, _ := time.Parse(time.RFC3339, ctx.Query("start_time"))
	endTime, _ := time.Parse(time.RFC3339, ctx.Query("end_time"))

	// If no explicit start_time, derive it from lookback_minutes (what the UI sends)
	if startTime.IsZero() {
		if mins, err := strconv.Atoi(ctx.Query("lookback_minutes")); err == nil && mins > 0 {
			startTime = time.Now().UTC().Add(-time.Duration(mins) * time.Minute)
		}
	}

	return observabilityrequests.Filters{
		ServiceID:   ctx.Query("service_id"),
		ServiceName: ctx.Query("service_name"),
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

func (c *Controller) ListTransactions(ctx *gin.Context) {
	rows, customError := c.service.ListTransactions(ctx, claimsFromContext(ctx), buildFilters(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func (c *Controller) ListErrorGroups(ctx *gin.Context) {
	rows, customError := c.service.ListErrorGroups(ctx, claimsFromContext(ctx), buildFilters(ctx))
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, rows)
}

func statusCode(code errs.Code) int {
	switch code {
	case "BAD_REQUEST":
		return http.StatusBadRequest
	case "NOT_FOUND":
		return http.StatusNotFound
	case "DEPENDENCY_UNAVAILABLE":
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func (c *Controller) GetServiceMap(ctx *gin.Context) {
	lookback, _ := strconv.Atoi(ctx.DefaultQuery("lookback_minutes", "60"))
	if lookback <= 0 {
		lookback = 60
	}
	topology, customError := c.service.GetServiceMap(ctx, claimsFromContext(ctx), lookback)
	if customError.Exists() {
		platformresponse.Error(ctx, statusCode(customError.Code()), customError)
		return
	}
	platformresponse.Success(ctx, topology)
}
