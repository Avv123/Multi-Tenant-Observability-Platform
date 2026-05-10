package cacheversion

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

const (
	ScopeTelemetryOverview = "telemetry_overview"
	ScopeLogs              = "logs"
	ScopeMetrics           = "metrics"
	ScopeTraces            = "traces"
	ScopeServiceHealth     = "service_health"
	ScopeLogAnalytics      = "analytics_logs"
	ScopeMetricAnalytics   = "analytics_metrics"
	ScopeTraceAnalytics    = "analytics_traces"
	ScopeSavedQueries      = "saved_queries"
	ScopeDashboards        = "dashboards"
)

func Current(ctx context.Context, client *redis.Client, tenantID string, scope string) string {
	if client == nil || strings.TrimSpace(tenantID) == "" {
		return "0"
	}
	value, err := client.Get(ctx, key(tenantID, scope)).Result()
	if err == nil && strings.TrimSpace(value) != "" {
		return value
	}
	return "0"
}

func Bump(ctx context.Context, client *redis.Client, tenantID string, scope string) {
	if client == nil || strings.TrimSpace(tenantID) == "" {
		return
	}
	_ = client.Incr(ctx, key(tenantID, scope)).Err()
}

func BumpMany(ctx context.Context, client *redis.Client, tenantID string, scopes ...string) {
	for _, scope := range scopes {
		Bump(ctx, client, tenantID, scope)
	}
}

func key(tenantID string, scope string) string {
	normalizedScope := strings.TrimSpace(scope)
	if normalizedScope == "" {
		normalizedScope = ScopeTelemetryOverview
	}
	return fmt.Sprintf("pulselens:cache-version:%s:%s", tenantID, normalizedScope)
}
