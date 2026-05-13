package services

import (
	"context"
	"encoding/json"

	commonauth "github.com/omniful/pulselens-common/auth"
	"github.com/omniful/pulselens-platform/cacheversion"
	"github.com/omniful/pulselens-platform/errs"
	observabilitymodels "github.com/omniful/pulselens-query-service/internal/observability/models"
	observabilityrepositories "github.com/omniful/pulselens-query-service/internal/observability/repositories"
	observabilityrequests "github.com/omniful/pulselens-query-service/internal/observability/requests"
	observabilityresponses "github.com/omniful/pulselens-query-service/internal/observability/responses"
	queryclickhouse "github.com/omniful/pulselens-query-service/pkg/clickhouse"
	"github.com/omniful/pulselens-query-service/pkg/postgres"
	queryredis "github.com/omniful/pulselens-query-service/pkg/redis"
)

type Service struct {
	repository *observabilityrepositories.Repository
}

func New() *Service {
	return &Service{repository: observabilityrepositories.NewRepository(postgres.Get(), queryclickhouse.Get())}
}

func (s *Service) Overview(ctx context.Context, claims *commonauth.Claims) (map[string]interface{}, errs.CustomError) {
	runtimeRows, _ := s.PlatformRuntime(ctx)
	backpressureRows, _ := s.PlatformBackpressure(ctx)
	cleanupRuns, _ := s.CleanupRuns(ctx, 5)
	logCount, err := cachedValueWithError(ctx, cacheKey(ctx, "overview:log_count", claims.TenantID, cacheversion.ScopeTelemetryOverview, map[string]any{"tenant_id": claims.TenantID}), func() (int64, error) {
		return s.repository.CountLogs(ctx, claims.TenantID)
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	metricCount, err := cachedValueWithError(ctx, cacheKey(ctx, "overview:metric_count", claims.TenantID, cacheversion.ScopeTelemetryOverview, map[string]any{"tenant_id": claims.TenantID}), func() (int64, error) {
		return s.repository.CountMetrics(ctx, claims.TenantID)
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	traceCount, err := cachedValueWithError(ctx, cacheKey(ctx, "overview:trace_count", claims.TenantID, cacheversion.ScopeTelemetryOverview, map[string]any{"tenant_id": claims.TenantID}), func() (int64, error) {
		return s.repository.CountTraceSpans(ctx, claims.TenantID)
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	serviceCount, err := cachedValueWithError(ctx, cacheKey(ctx, "overview:service_count", claims.TenantID, cacheversion.ScopeTelemetryOverview, map[string]any{"tenant_id": claims.TenantID}), func() (int64, error) {
		return s.repository.CountServices(ctx, claims.TenantID)
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	serviceHealth, customError := s.ServiceHealth(ctx, claims, 20)
	if customError.Exists() {
		return nil, customError
	}
	latestLogs, customError := s.ListLogs(ctx, claims, observabilityrequests.Filters{Limit: 10})
	if customError.Exists() {
		return nil, customError
	}
	latestMetrics, customError := s.ListMetrics(ctx, claims, observabilityrequests.Filters{Limit: 10})
	if customError.Exists() {
		return nil, customError
	}
	return map[string]interface{}{
		"tenant_id":      claims.TenantID,
		"log_count":      logCount,
		"metric_count":   metricCount,
		"trace_count":    traceCount,
		"service_count":  serviceCount,
		"service_health": serviceHealth,
		"latest_logs":    latestLogs,
		"latest_metrics": latestMetrics,
		"usage":          s.repository.ListUsage(ctx, claims.TenantID, 20),
		"runtime":        runtimeRows,
		"backpressure":   backpressureRows,
		"cleanup_runs":   cleanupRuns,
	}, errs.CustomError{}
}

func (s *Service) ListLogs(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) ([]interface{}, errs.CustomError) {
	rows, err := cachedValueWithError(ctx, cacheKey(ctx, "logs", claims.TenantID, cacheversion.ScopeLogs, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() ([]interface{}, error) {
		typedRows, listErr := s.repository.ListLogs(ctx, claims.TenantID, filters)
		if listErr != nil {
			return nil, listErr
		}
		result := make([]interface{}, 0, len(typedRows))
		for _, row := range typedRows {
			result = append(result, row)
		}
		return result, nil
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) ListMetrics(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) ([]interface{}, errs.CustomError) {
	rows, err := cachedValueWithError(ctx, cacheKey(ctx, "metrics", claims.TenantID, cacheversion.ScopeMetrics, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() ([]interface{}, error) {
		typedRows, listErr := s.repository.ListMetrics(ctx, claims.TenantID, filters)
		if listErr != nil {
			return nil, listErr
		}
		result := make([]interface{}, 0, len(typedRows))
		for _, row := range typedRows {
			result = append(result, row)
		}
		return result, nil
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) ListTraces(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) ([]interface{}, errs.CustomError) {
	rows, err := cachedValueWithError(ctx, cacheKey(ctx, "traces", claims.TenantID, cacheversion.ScopeTraces, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() ([]interface{}, error) {
		typedRows, listErr := s.repository.ListTraces(ctx, claims.TenantID, filters)
		if listErr != nil {
			return nil, listErr
		}
		result := make([]interface{}, 0, len(typedRows))
		for _, row := range typedRows {
			result = append(result, row)
		}
		return result, nil
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) TraceDetail(ctx context.Context, claims *commonauth.Claims, traceID string) ([]interface{}, errs.CustomError) {
	rows, err := cachedValueWithError(ctx, cacheKey(ctx, "trace_detail", claims.TenantID, cacheversion.ScopeTraces, map[string]any{"tenant_id": claims.TenantID, "trace_id": traceID}), func() ([]interface{}, error) {
		typedRows, detailErr := s.repository.TraceDetail(ctx, claims.TenantID, traceID)
		if detailErr != nil {
			return nil, detailErr
		}
		result := make([]interface{}, 0, len(typedRows))
		for _, row := range typedRows {
			result = append(result, row)
		}
		return result, nil
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) ServiceHealth(ctx context.Context, claims *commonauth.Claims, limit int) ([]interface{}, errs.CustomError) {
	rows, err := cachedValueWithError(ctx, cacheKey(ctx, "service_health", claims.TenantID, cacheversion.ScopeServiceHealth, map[string]any{"tenant_id": claims.TenantID, "limit": limit}), func() ([]interface{}, error) {
		typedRows, healthErr := s.repository.ListServiceHealth(ctx, claims.TenantID, limit)
		if healthErr != nil {
			return nil, healthErr
		}
		result := make([]interface{}, 0, len(typedRows))
		for _, row := range typedRows {
			result = append(result, row)
		}
		return result, nil
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) CreateSavedQuery(ctx context.Context, claims *commonauth.Claims, request observabilityrequests.CreateSavedQueryRequest) (observabilitymodels.SavedQuery, error) {
	row, err := s.repository.CreateSavedQuery(ctx, claims, request)
	if err == nil {
		cacheversion.Bump(ctx, queryredis.Get(), claims.TenantID, cacheversion.ScopeSavedQueries)
	}
	return row, err
}

func (s *Service) ListSavedQueries(ctx context.Context, claims *commonauth.Claims) ([]observabilitymodels.SavedQuery, error) {
	return cachedValue(ctx, cacheKey(ctx, "saved_queries", claims.TenantID, cacheversion.ScopeSavedQueries, map[string]any{"tenant_id": claims.TenantID}), func() []observabilitymodels.SavedQuery {
		rows, _ := s.repository.ListSavedQueries(ctx, claims.TenantID)
		return rows
	}), nil
}

func (s *Service) UpdateSavedQuery(ctx context.Context, claims *commonauth.Claims, queryID string, request observabilityrequests.UpdateSavedQueryRequest) (observabilitymodels.SavedQuery, error) {
	row, err := s.repository.UpdateSavedQuery(ctx, claims, queryID, request)
	if err == nil {
		cacheversion.Bump(ctx, queryredis.Get(), claims.TenantID, cacheversion.ScopeSavedQueries)
	}
	return row, err
}

func (s *Service) CreateDashboard(ctx context.Context, claims *commonauth.Claims, request observabilityrequests.CreateDashboardRequest) (observabilityresponses.Dashboard, error) {
	row, err := s.repository.CreateDashboard(ctx, claims, createToUpdateDashboardRequest(request))
	if err == nil {
		cacheversion.Bump(ctx, queryredis.Get(), claims.TenantID, cacheversion.ScopeDashboards)
	}
	return buildDashboardResponse(row), err
}

func (s *Service) ListDashboards(ctx context.Context, claims *commonauth.Claims) ([]observabilityresponses.Dashboard, error) {
	return cachedValue(ctx, cacheKey(ctx, "dashboards", claims.TenantID, cacheversion.ScopeDashboards, map[string]any{"tenant_id": claims.TenantID}), func() []observabilityresponses.Dashboard {
		rows, _ := s.repository.ListDashboards(ctx, claims.TenantID)
		result := make([]observabilityresponses.Dashboard, 0, len(rows))
		for _, row := range rows {
			result = append(result, buildDashboardResponse(row))
		}
		return result
	}), nil
}

func (s *Service) UpdateDashboard(ctx context.Context, claims *commonauth.Claims, dashboardID string, request observabilityrequests.UpdateDashboardRequest) (observabilityresponses.Dashboard, error) {
	row, err := s.repository.UpdateDashboard(ctx, claims, dashboardID, normalizeDashboardRequest(request))
	if err == nil {
		cacheversion.Bump(ctx, queryredis.Get(), claims.TenantID, cacheversion.ScopeDashboards)
	}
	return buildDashboardResponse(row), err
}

func (s *Service) UpdateDashboardWidget(ctx context.Context, claims *commonauth.Claims, dashboardID string, widgetID string, request observabilityrequests.UpdateDashboardWidgetRequest) (observabilityresponses.Dashboard, error) {
	row, err := s.repository.GetDashboard(ctx, claims.TenantID, dashboardID)
	if err != nil {
		return observabilityresponses.Dashboard{}, err
	}
	parsed := createToUpdateDashboardRequest(observabilityrequests.CreateDashboardRequest{
		Name:             row.Name,
		Description:      row.Description,
		DefaultTimeRange: row.DefaultTimeRange,
		Layout:           map[string]any{},
	})
	_ = json.Unmarshal([]byte(row.Layout), &parsed.Layout)
	_ = json.Unmarshal([]byte(row.Widgets), &parsed.Widgets)
	found := false
	for index := range parsed.Widgets {
		if parsed.Widgets[index].ID != widgetID {
			continue
		}
		parsed.Widgets[index].Title = request.Title
		parsed.Widgets[index].Type = request.Type
		parsed.Widgets[index].Dataset = request.Dataset
		parsed.Widgets[index].ChartType = request.ChartType
		parsed.Widgets[index].Metric = request.Metric
		parsed.Widgets[index].Filters = request.Filters
		parsed.Widgets[index].Layout = request.Layout
		found = true
		break
	}
	if !found {
		return observabilityresponses.Dashboard{}, errs.New("NOT_FOUND", "widget not found")
	}
	return s.UpdateDashboard(ctx, claims, dashboardID, parsed)
}

func (s *Service) DeleteDashboardWidget(ctx context.Context, claims *commonauth.Claims, dashboardID string, widgetID string) error {
	row, err := s.repository.GetDashboard(ctx, claims.TenantID, dashboardID)
	if err != nil {
		return err
	}
	parsed := createToUpdateDashboardRequest(observabilityrequests.CreateDashboardRequest{
		Name:             row.Name,
		Description:      row.Description,
		DefaultTimeRange: row.DefaultTimeRange,
		Layout:           map[string]any{},
	})
	_ = json.Unmarshal([]byte(row.Layout), &parsed.Layout)
	_ = json.Unmarshal([]byte(row.Widgets), &parsed.Widgets)
	filtered := make([]observabilityrequests.DashboardWidget, 0, len(parsed.Widgets))
	found := false
	for _, widget := range parsed.Widgets {
		if widget.ID == widgetID {
			found = true
			continue
		}
		filtered = append(filtered, widget)
	}
	if !found {
		return errs.New("NOT_FOUND", "widget not found")
	}
	parsed.Widgets = filtered
	_, err = s.UpdateDashboard(ctx, claims, dashboardID, parsed)
	return err
}

func (s *Service) LogSeveritySeries(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) (interface{}, errs.CustomError) {
	rows, err := cachedValueWithError(ctx, cacheKey(ctx, "analytics:log_severity", claims.TenantID, cacheversion.ScopeLogAnalytics, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() (interface{}, error) {
		return s.repository.ListLogSeverityRollups(ctx, claims.TenantID, filters)
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) MetricSeries(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) (interface{}, errs.CustomError) {
	rows, err := cachedValueWithError(ctx, cacheKey(ctx, "analytics:metric_series", claims.TenantID, cacheversion.ScopeMetricAnalytics, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() (interface{}, error) {
		return s.repository.ListMetricSeries(ctx, claims.TenantID, filters)
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	return rows, errs.CustomError{}
}

func (s *Service) TraceLatencySeries(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) (interface{}, errs.CustomError) {
	rows, err := cachedValueWithError(ctx, cacheKey(ctx, "analytics:trace_latency", claims.TenantID, cacheversion.ScopeTraceAnalytics, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() (interface{}, error) {
		return s.repository.ListTraceLatencyRollups(ctx, claims.TenantID, filters)
	})
	if err != nil {
		return nil, telemetryUnavailableError(err)
	}
	return rows, errs.CustomError{}
}
