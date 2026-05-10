package services

import (
	"context"

	commonauth "github.com/omniful/pulselens-common/auth"
	"github.com/omniful/pulselens-platform/cacheversion"
	observabilitymodels "github.com/omniful/pulselens-query-service/internal/observability/models"
	observabilityrepositories "github.com/omniful/pulselens-query-service/internal/observability/repositories"
	observabilityrequests "github.com/omniful/pulselens-query-service/internal/observability/requests"
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

func (s *Service) Overview(ctx context.Context, claims *commonauth.Claims) map[string]interface{} {
	runtimeRows, _ := s.PlatformRuntime(ctx)
	backpressureRows, _ := s.PlatformBackpressure(ctx)
	cleanupRuns, _ := s.CleanupRuns(ctx, 5)
	logCount := cachedValue(ctx, cacheKey(ctx, "overview:log_count", claims.TenantID, cacheversion.ScopeTelemetryOverview, map[string]any{"tenant_id": claims.TenantID}), func() int64 {
		return s.repository.CountLogs(ctx, claims.TenantID)
	})
	metricCount := cachedValue(ctx, cacheKey(ctx, "overview:metric_count", claims.TenantID, cacheversion.ScopeTelemetryOverview, map[string]any{"tenant_id": claims.TenantID}), func() int64 {
		return s.repository.CountMetrics(ctx, claims.TenantID)
	})
	traceCount := cachedValue(ctx, cacheKey(ctx, "overview:trace_count", claims.TenantID, cacheversion.ScopeTelemetryOverview, map[string]any{"tenant_id": claims.TenantID}), func() int64 {
		return s.repository.CountTraceSpans(ctx, claims.TenantID)
	})
	serviceCount := cachedValue(ctx, cacheKey(ctx, "overview:service_count", claims.TenantID, cacheversion.ScopeTelemetryOverview, map[string]any{"tenant_id": claims.TenantID}), func() int64 {
		return s.repository.CountServices(ctx, claims.TenantID)
	})
	serviceHealth := s.ServiceHealth(ctx, claims, 20)
	latestLogs := s.ListLogs(ctx, claims, observabilityrequests.Filters{Limit: 10})
	latestMetrics := s.ListMetrics(ctx, claims, observabilityrequests.Filters{Limit: 10})
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
	}
}

func (s *Service) ListLogs(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) interface{} {
	return cachedValue(ctx, cacheKey(ctx, "logs", claims.TenantID, cacheversion.ScopeLogs, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() []interface{} {
		rows := s.repository.ListLogs(ctx, claims.TenantID, filters)
		result := make([]interface{}, 0, len(rows))
		for _, row := range rows {
			result = append(result, row)
		}
		return result
	})
}

func (s *Service) ListMetrics(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) interface{} {
	return cachedValue(ctx, cacheKey(ctx, "metrics", claims.TenantID, cacheversion.ScopeMetrics, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() []interface{} {
		rows := s.repository.ListMetrics(ctx, claims.TenantID, filters)
		result := make([]interface{}, 0, len(rows))
		for _, row := range rows {
			result = append(result, row)
		}
		return result
	})
}

func (s *Service) ListTraces(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) interface{} {
	return cachedValue(ctx, cacheKey(ctx, "traces", claims.TenantID, cacheversion.ScopeTraces, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() []interface{} {
		rows := s.repository.ListTraces(ctx, claims.TenantID, filters)
		result := make([]interface{}, 0, len(rows))
		for _, row := range rows {
			result = append(result, row)
		}
		return result
	})
}

func (s *Service) TraceDetail(ctx context.Context, claims *commonauth.Claims, traceID string) interface{} {
	return cachedValue(ctx, cacheKey(ctx, "trace_detail", claims.TenantID, cacheversion.ScopeTraces, map[string]any{"tenant_id": claims.TenantID, "trace_id": traceID}), func() []interface{} {
		rows := s.repository.TraceDetail(ctx, claims.TenantID, traceID)
		result := make([]interface{}, 0, len(rows))
		for _, row := range rows {
			result = append(result, row)
		}
		return result
	})
}

func (s *Service) ServiceHealth(ctx context.Context, claims *commonauth.Claims, limit int) interface{} {
	return cachedValue(ctx, cacheKey(ctx, "service_health", claims.TenantID, cacheversion.ScopeServiceHealth, map[string]any{"tenant_id": claims.TenantID, "limit": limit}), func() []interface{} {
		rows := s.repository.ListServiceHealth(ctx, claims.TenantID, limit)
		result := make([]interface{}, 0, len(rows))
		for _, row := range rows {
			result = append(result, row)
		}
		return result
	})
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

func (s *Service) CreateDashboard(ctx context.Context, claims *commonauth.Claims, request observabilityrequests.CreateDashboardRequest) (observabilitymodels.Dashboard, error) {
	row, err := s.repository.CreateDashboard(ctx, claims, request)
	if err == nil {
		cacheversion.Bump(ctx, queryredis.Get(), claims.TenantID, cacheversion.ScopeDashboards)
	}
	return row, err
}

func (s *Service) ListDashboards(ctx context.Context, claims *commonauth.Claims) ([]observabilitymodels.Dashboard, error) {
	return cachedValue(ctx, cacheKey(ctx, "dashboards", claims.TenantID, cacheversion.ScopeDashboards, map[string]any{"tenant_id": claims.TenantID}), func() []observabilitymodels.Dashboard {
		rows, _ := s.repository.ListDashboards(ctx, claims.TenantID)
		return rows
	}), nil
}

func (s *Service) UpdateDashboard(ctx context.Context, claims *commonauth.Claims, dashboardID string, request observabilityrequests.UpdateDashboardRequest) (observabilitymodels.Dashboard, error) {
	row, err := s.repository.UpdateDashboard(ctx, claims, dashboardID, request)
	if err == nil {
		cacheversion.Bump(ctx, queryredis.Get(), claims.TenantID, cacheversion.ScopeDashboards)
	}
	return row, err
}

func (s *Service) LogSeveritySeries(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) interface{} {
	return cachedValue(ctx, cacheKey(ctx, "analytics:log_severity", claims.TenantID, cacheversion.ScopeLogAnalytics, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() interface{} {
		return s.repository.ListLogSeverityRollups(ctx, claims.TenantID, filters)
	})
}

func (s *Service) MetricSeries(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) interface{} {
	return cachedValue(ctx, cacheKey(ctx, "analytics:metric_series", claims.TenantID, cacheversion.ScopeMetricAnalytics, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() interface{} {
		return s.repository.ListMetricSeries(ctx, claims.TenantID, filters)
	})
}

func (s *Service) TraceLatencySeries(ctx context.Context, claims *commonauth.Claims, filters observabilityrequests.Filters) interface{} {
	return cachedValue(ctx, cacheKey(ctx, "analytics:trace_latency", claims.TenantID, cacheversion.ScopeTraceAnalytics, map[string]any{"tenant_id": claims.TenantID, "filters": filters}), func() interface{} {
		return s.repository.ListTraceLatencyRollups(ctx, claims.TenantID, filters)
	})
}
