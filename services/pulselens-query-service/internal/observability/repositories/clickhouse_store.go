package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	platformclickhouse "github.com/Avv123/pulselens-platform/clickhouse"
	observabilityrequests "github.com/Avv123/pulselens-query-service/internal/observability/requests"
	observabilityresponses "github.com/Avv123/pulselens-query-service/internal/observability/responses"
)

type countRow struct {
	Count int64 `json:"count"`
}

type clickhouseLogRow struct {
	EventID      string `json:"event_id"`
	ServiceName  string `json:"service_name"`
	Environment  string `json:"environment"`
	Severity     string `json:"severity"`
	Message      string `json:"message"`
	TraceID      string `json:"trace_id"`
	Payload      string `json:"payload"`
	OccurredAtMS int64  `json:"occurred_at_ms"`
}

type clickhouseMetricRow struct {
	EventID      string  `json:"event_id"`
	ServiceName  string  `json:"service_name"`
	Environment  string  `json:"environment"`
	MetricName   string  `json:"metric_name"`
	Value        float64 `json:"value"`
	OccurredAtMS int64   `json:"occurred_at_ms"`
}

type clickhouseTraceRow struct {
	TraceID       string `json:"trace_id"`
	ServiceName   string `json:"service_name"`
	Environment   string `json:"environment"`
	SpanCount     int64  `json:"span_count"`
	FirstSeenAtMS int64  `json:"first_seen_at_ms"`
	LastSeenAtMS  int64  `json:"last_seen_at_ms"`
}

type clickhouseTraceSpanRow struct {
	EventID      string `json:"event_id"`
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id"`
	ParentSpanID string `json:"parent_span_id"`
	Operation    string `json:"operation"`
	Status       string `json:"status"`
	ServiceName  string `json:"service_name"`
	Environment  string `json:"environment"`
	DurationMS   int64  `json:"duration_ms"`
	Payload      string `json:"payload"`
	OccurredAtMS int64  `json:"occurred_at_ms"`
}

type clickhouseTransactionRow struct {
	ServiceName     string `json:"service_name"`
	Operation       string `json:"operation"`
	TotalCalls      uint64 `json:"total_calls"`
	TotalErrors     uint64 `json:"total_errors"`
	TotalDurationMS int64  `json:"sum_duration_ms"` // alias differs from column to avoid CH 24.x nested-agg error
	MaxDurationMS   int64  `json:"max_duration_ms"`
}

type clickhouseErrorGroupRow struct {
	ServiceName  string `json:"service_name"`
	Severity     string `json:"severity"`
	Title        string `json:"title"`
	Occurrences  int64  `json:"occurrences"`
	FirstSeenMS  int64  `json:"first_seen_at_ms"`
	LastSeenMS   int64  `json:"last_seen_at_ms"`
	SampleTraces string `json:"sample_trace_ids"`
}

type clickhouseServiceHealthRow struct {
	ServiceID        string `json:"service_id"`
	ServiceName      string `json:"service_name"`
	Environment      string `json:"environment"`
	LastEventAtMS    int64  `json:"last_event_at_ms"`
	EventCount       int64  `json:"event_count"`
	ErrorLogCount    int64  `json:"error_log_count"`
	CriticalLogCount int64  `json:"critical_log_count"`
	LatestMetricAtMS int64  `json:"latest_metric_at_ms"`
	LatestTraceAtMS  int64  `json:"latest_trace_at_ms"`
}

type clickhouseLogSeverityRollupRow struct {
	ServiceID     string `json:"service_id"`
	ServiceName   string `json:"service_name"`
	Environment   string `json:"environment"`
	Severity      string `json:"severity"`
	BucketStartMS int64  `json:"bucket_start_ms"`
	EventCount    int64  `json:"event_count"`
}

type clickhouseMetricRollupRow struct {
	ServiceID     string  `json:"service_id"`
	Environment   string  `json:"environment"`
	MetricName    string  `json:"metric_name"`
	BucketStartMS int64   `json:"bucket_start_ms"`
	SampleCount   int64   `json:"sample_count"`
	SumValue      float64 `json:"sum_value"`
	MinValue      float64 `json:"min_value"`
	MaxValue      float64 `json:"max_value"`
	LastValue     float64 `json:"last_value"`
}

type clickhouseTraceLatencyRollupRow struct {
	ServiceID       string `json:"service_id"`
	ServiceName     string `json:"service_name"`
	Environment     string `json:"environment"`
	Operation       string `json:"operation"`
	BucketStartMS   int64  `json:"bucket_start_ms"`
	SpanCount       int64  `json:"span_count"`
	ErrorCount      int64  `json:"error_count"`
	TotalDurationMS int64  `json:"total_duration_ms"`
	MaxDurationMS   int64  `json:"max_duration_ms"`
}

func (r *Repository) countLogsCH(ctx context.Context, tenantID string) (int64, error) {
	return r.countFromClickHouse(ctx, fmt.Sprintf("SELECT coalesce(sum(event_count), 0) AS count FROM pulselens.telemetry_rollup_minutes WHERE tenant_id = '%s' AND event_type = 'log'", escapeLiteral(tenantID)))
}

func (r *Repository) countMetricsCH(ctx context.Context, tenantID string) (int64, error) {
	return r.countFromClickHouse(ctx, fmt.Sprintf("SELECT coalesce(sum(event_count), 0) AS count FROM pulselens.telemetry_rollup_minutes WHERE tenant_id = '%s' AND event_type = 'metric'", escapeLiteral(tenantID)))
}

func (r *Repository) countTraceSpansCH(ctx context.Context, tenantID string) (int64, error) {
	return r.countFromClickHouse(ctx, fmt.Sprintf("SELECT coalesce(sum(event_count), 0) AS count FROM pulselens.telemetry_rollup_minutes WHERE tenant_id = '%s' AND event_type = 'trace'", escapeLiteral(tenantID)))
}

func (r *Repository) countServicesCH(ctx context.Context, tenantID string) (int64, error) {
	return r.countFromClickHouse(ctx, fmt.Sprintf("SELECT uniq(service_id) AS count FROM pulselens.service_health_rollup_minutes WHERE tenant_id = '%s'", escapeLiteral(tenantID)))
}

func (r *Repository) listLogsCH(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.LogRow, error) {
	query := fmt.Sprintf(`
		SELECT event_id, service_name, environment, severity, message, trace_id, payload,
		       toUnixTimestamp64Milli(occurred_at) AS occurred_at_ms
		FROM pulselens.log_events
		WHERE tenant_id = '%s'%s
		ORDER BY occurred_at DESC
		LIMIT %d OFFSET %d
	`, escapeLiteral(tenantID), logFilterClause(filters), filters.Limit, filters.Offset)

	rows, err := platformclickhouse.Select[clickhouseLogRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}

	result := make([]observabilityresponses.LogRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, observabilityresponses.LogRow{
			EventID:     row.EventID,
			ServiceName: row.ServiceName,
			Environment: row.Environment,
			Severity:    row.Severity,
			Message:     row.Message,
			TraceID:     row.TraceID,
			Payload:     row.Payload,
			OccurredAt:  fromUnixMillis(row.OccurredAtMS),
		})
	}
	return result, nil
}

func (r *Repository) listMetricsCH(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.MetricRow, error) {
	query := fmt.Sprintf(`
		SELECT event_id, service_name, environment, metric_name, value,
		       toUnixTimestamp64Milli(occurred_at) AS occurred_at_ms
		FROM pulselens.metric_points
		WHERE tenant_id = '%s'%s
		ORDER BY occurred_at DESC
		LIMIT %d OFFSET %d
	`, escapeLiteral(tenantID), metricFilterClause(filters), filters.Limit, filters.Offset)

	rows, err := platformclickhouse.Select[clickhouseMetricRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}

	result := make([]observabilityresponses.MetricRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, observabilityresponses.MetricRow{
			EventID:     row.EventID,
			ServiceName: row.ServiceName,
			Environment: row.Environment,
			MetricName:  row.MetricName,
			Value:       row.Value,
			OccurredAt:  fromUnixMillis(row.OccurredAtMS),
		})
	}
	return result, nil
}

func (r *Repository) listTracesCH(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.TraceRow, error) {
	query := fmt.Sprintf(`
		SELECT trace_id, any(service_name) AS service_name, any(environment) AS environment,
		       count() AS span_count,
		       toUnixTimestamp64Milli(min(occurred_at)) AS first_seen_at_ms,
		       toUnixTimestamp64Milli(max(occurred_at)) AS last_seen_at_ms
		FROM pulselens.trace_spans
		WHERE tenant_id = '%s'%s
		GROUP BY trace_id
		ORDER BY last_seen_at_ms DESC
		LIMIT %d OFFSET %d
	`, escapeLiteral(tenantID), traceFilterClause(filters), filters.Limit, filters.Offset)

	rows, err := platformclickhouse.Select[clickhouseTraceRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}

	result := make([]observabilityresponses.TraceRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, observabilityresponses.TraceRow{
			TraceID:     row.TraceID,
			ServiceName: row.ServiceName,
			Environment: row.Environment,
			SpanCount:   row.SpanCount,
			FirstSeenAt: fromUnixMillis(row.FirstSeenAtMS),
			LastSeenAt:  fromUnixMillis(row.LastSeenAtMS),
		})
	}
	return result, nil
}

func (r *Repository) traceDetailCH(ctx context.Context, tenantID string, traceID string) ([]observabilityresponses.TraceSpanRow, error) {
	query := fmt.Sprintf(`
		SELECT event_id, trace_id, span_id, parent_span_id, operation, status, service_name, environment,
		       duration_ms, payload,
		       toUnixTimestamp64Milli(occurred_at) AS occurred_at_ms
		FROM pulselens.trace_spans
		WHERE tenant_id = '%s' AND trace_id = '%s'
		ORDER BY occurred_at ASC
	`, escapeLiteral(tenantID), escapeLiteral(traceID))

	rows, err := platformclickhouse.Select[clickhouseTraceSpanRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}

	result := make([]observabilityresponses.TraceSpanRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, observabilityresponses.TraceSpanRow{
			EventID:      row.EventID,
			TraceID:      row.TraceID,
			SpanID:       row.SpanID,
			ParentSpanID: row.ParentSpanID,
			Operation:    row.Operation,
			Status:       row.Status,
			ServiceName:  row.ServiceName,
			Environment:  row.Environment,
			DurationMS:   row.DurationMS,
			Payload:      row.Payload,
			OccurredAt:   fromUnixMillis(row.OccurredAtMS),
		})
	}
	return result, nil
}

func (r *Repository) listTransactionsCH(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.TransactionRow, error) {
	serviceClause := ""
	if filters.ServiceName != "" {
		serviceClause = fmt.Sprintf(" AND lower(service_name) LIKE lower('%%%s%%')", escapeLiteral(filters.ServiceName))
	}
	timeClause := ""
	if !filters.StartTime.IsZero() {
		timeClause += fmt.Sprintf(" AND bucket_start >= toDateTime64('%s', 3, 'UTC')", filters.StartTime.UTC().Format("2006-01-02 15:04:05.000"))
	}
	// Each row in trace_latency_rollup_minutes represents exactly one span (span_count=1).
	// Use count() / countIf() to avoid ILLEGAL_AGGREGATION in ClickHouse 24.x.
	// Use a different alias (sum_duration_ms) so it doesn't clash with the source column
	// name total_duration_ms, which would confuse the query planner.
	query := fmt.Sprintf(`
		SELECT
			any(service_name)          AS service_name,
			operation,
			count()                    AS total_calls,
			countIf(error_count > 0)   AS total_errors,
			sum(total_duration_ms)     AS sum_duration_ms,
			max(max_duration_ms)       AS max_duration_ms
		FROM pulselens.trace_latency_rollup_minutes
		WHERE tenant_id = '%s'%s%s
		GROUP BY operation
		ORDER BY count() DESC
		LIMIT %d
	`, escapeLiteral(tenantID), serviceClause, timeClause, filters.Limit)

	rows, err := platformclickhouse.Select[clickhouseTransactionRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}

	result := make([]observabilityresponses.TransactionRow, 0, len(rows))
	for _, row := range rows {
		var avgMS float64
		if row.TotalCalls > 0 {
			avgMS = float64(row.TotalDurationMS) / float64(row.TotalCalls)
		}
		var errorPct float64
		if row.TotalCalls > 0 {
			errorPct = float64(row.TotalErrors) / float64(row.TotalCalls) * 100
		}
		result = append(result, observabilityresponses.TransactionRow{
			ServiceName:   row.ServiceName,
			Operation:     row.Operation,
			TotalCalls:    int64(row.TotalCalls),
			TotalErrors:   int64(row.TotalErrors),
			AvgDurationMS: avgMS,
			MaxDurationMS: row.MaxDurationMS,
			ErrorPct:      errorPct,
		})
	}
	return result, nil
}

func (r *Repository) listErrorGroupsCH(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.ErrorGroupRow, error) {
	serviceClause := ""
	if filters.ServiceName != "" {
		serviceClause = fmt.Sprintf(" AND lower(service_name) LIKE lower('%%%s%%')", escapeLiteral(filters.ServiceName))
	}
	lookback := "7 DAY"
	if !filters.StartTime.IsZero() {
		lookback = fmt.Sprintf("0 SECOND' AND occurred_at >= toDateTime64('%s', 3, 'UTC') -- ", filters.StartTime.UTC().Format("2006-01-02 15:04:05.000"))
	}
	query := fmt.Sprintf(`
		SELECT
			service_name,
			severity,
			substring(message, 1, 120)                       AS title,
			count()                                          AS occurrences,
			toUnixTimestamp64Milli(min(occurred_at))         AS first_seen_at_ms,
			toUnixTimestamp64Milli(max(occurred_at))         AS last_seen_at_ms,
			arrayStringConcat(groupUniqArray(5)(trace_id), ',') AS sample_trace_ids
		FROM pulselens.log_events
		WHERE tenant_id = '%s'
		  AND severity IN ('error', 'critical', 'warn')%s
		  AND occurred_at >= now() - INTERVAL %s
		GROUP BY service_name, severity, title
		ORDER BY occurrences DESC
		LIMIT %d
	`, escapeLiteral(tenantID), serviceClause, lookback, filters.Limit)

	rows, err := platformclickhouse.Select[clickhouseErrorGroupRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}

	result := make([]observabilityresponses.ErrorGroupRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, observabilityresponses.ErrorGroupRow{
			ServiceName:    row.ServiceName,
			Severity:       row.Severity,
			Title:          row.Title,
			Occurrences:    row.Occurrences,
			FirstSeenAt:    fromUnixMillis(row.FirstSeenMS),
			LastSeenAt:     fromUnixMillis(row.LastSeenMS),
			SampleTraceIDs: strings.Split(row.SampleTraces, ","),
		})
	}
	return result, nil
}

func (r *Repository) listServiceHealthCH(ctx context.Context, tenantID string, limit int) ([]observabilityresponses.ServiceHealthRow, error) {
	query := fmt.Sprintf(`
		SELECT
			service_id,
			any(service_name) AS service_name,
			any(environment) AS environment,
			toUnixTimestamp64Milli(max(last_event_at)) AS last_event_at_ms,
			sum(event_count) AS event_count,
			sum(error_log_count) AS error_log_count,
			sum(critical_log_count) AS critical_log_count,
			toUnixTimestamp64Milli(maxOrNull(latest_metric_at)) AS latest_metric_at_ms,
			toUnixTimestamp64Milli(maxOrNull(latest_trace_at)) AS latest_trace_at_ms
		FROM pulselens.service_health_rollup_minutes
		WHERE tenant_id = '%s'
		GROUP BY service_id
		ORDER BY last_event_at_ms DESC
		LIMIT %d
	`, escapeLiteral(tenantID), limit)

	rows, err := platformclickhouse.Select[clickhouseServiceHealthRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}

	result := make([]observabilityresponses.ServiceHealthRow, 0, len(rows))
	for _, row := range rows {
		converted := observabilityresponses.ServiceHealthRow{
			ServiceID:      row.ServiceID,
			ServiceName:    row.ServiceName,
			Environment:    row.Environment,
			LastEventAt:    fromUnixMillis(row.LastEventAtMS),
			EventCount:     row.EventCount,
			ErrorLogCount:  row.ErrorLogCount,
			CriticalCount:  row.CriticalLogCount,
			LatestMetricAt: fromUnixMillis(row.LatestMetricAtMS),
			LatestTraceAt:  fromUnixMillis(row.LatestTraceAtMS),
		}
		converted.HealthStatus = deriveHealth(converted)
		result = append(result, converted)
	}
	return result, nil
}

func (r *Repository) listLogSeverityRollupsCH(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.LogSeverityRollupRow, error) {
	query := fmt.Sprintf(`
		SELECT
			service_id,
			any(service_name) AS service_name,
			environment,
			severity,
			toUnixTimestamp64Milli(bucket_start) AS bucket_start_ms,
			sum(event_count) AS event_count
		FROM pulselens.log_severity_rollup_minutes
		WHERE tenant_id = '%s'%s
		GROUP BY service_id, environment, severity, bucket_start
		ORDER BY bucket_start DESC, service_name ASC, severity ASC
		LIMIT %d OFFSET %d
	`, escapeLiteral(tenantID), logRollupFilterClause(filters), filters.Limit, filters.Offset)
	rows, err := platformclickhouse.Select[clickhouseLogSeverityRollupRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}
	result := make([]observabilityresponses.LogSeverityRollupRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, observabilityresponses.LogSeverityRollupRow{
			ServiceID:   row.ServiceID,
			ServiceName: row.ServiceName,
			Environment: row.Environment,
			Severity:    row.Severity,
			BucketStart: fromUnixMillis(row.BucketStartMS),
			EventCount:  row.EventCount,
		})
	}
	return result, nil
}

func (r *Repository) listMetricSeriesCH(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.MetricSeriesRow, error) {
	query := fmt.Sprintf(`
		SELECT
			service_id,
			environment,
			metric_name,
			toUnixTimestamp64Milli(bucket_start) AS bucket_start_ms,
			sum(sample_count) AS sample_count,
			sum(sum_value) AS sum_value,
			min(min_value) AS min_value,
			max(max_value) AS max_value,
			argMax(last_value, last_event_at) AS last_value
		FROM pulselens.metric_rollup_minutes
		WHERE tenant_id = '%s'%s
		GROUP BY service_id, environment, metric_name, bucket_start
		ORDER BY bucket_start DESC, metric_name ASC
		LIMIT %d OFFSET %d
	`, escapeLiteral(tenantID), metricRollupFilterClause(filters), filters.Limit, filters.Offset)
	rows, err := platformclickhouse.Select[clickhouseMetricRollupRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}
	result := make([]observabilityresponses.MetricSeriesRow, 0, len(rows))
	for _, row := range rows {
		item := observabilityresponses.MetricSeriesRow{
			ServiceID:   row.ServiceID,
			Environment: row.Environment,
			MetricName:  row.MetricName,
			BucketStart: fromUnixMillis(row.BucketStartMS),
			SampleCount: row.SampleCount,
			SumValue:    row.SumValue,
			MinValue:    row.MinValue,
			MaxValue:    row.MaxValue,
			LastValue:   row.LastValue,
		}
		if item.SampleCount > 0 {
			item.AverageValue = item.SumValue / float64(item.SampleCount)
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *Repository) listTraceLatencyRollupsCH(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.TraceLatencyRollupRow, error) {
	query := fmt.Sprintf(`
		SELECT
			service_id,
			any(service_name) AS service_name,
			environment,
			operation,
			toUnixTimestamp64Milli(bucket_start) AS bucket_start_ms,
			sum(span_count) AS span_count,
			sum(error_count) AS error_count,
			sum(total_duration_ms) AS total_duration_ms,
			max(max_duration_ms) AS max_duration_ms
		FROM pulselens.trace_latency_rollup_minutes
		WHERE tenant_id = '%s'%s
		GROUP BY service_id, environment, operation, bucket_start
		ORDER BY bucket_start DESC, service_name ASC, operation ASC
		LIMIT %d OFFSET %d
	`, escapeLiteral(tenantID), traceRollupFilterClause(filters), filters.Limit, filters.Offset)
	rows, err := platformclickhouse.Select[clickhouseTraceLatencyRollupRow](ctx, r.ch, query)
	if err != nil {
		return nil, err
	}
	result := make([]observabilityresponses.TraceLatencyRollupRow, 0, len(rows))
	for _, row := range rows {
		item := observabilityresponses.TraceLatencyRollupRow{
			ServiceID:       row.ServiceID,
			ServiceName:     row.ServiceName,
			Environment:     row.Environment,
			Operation:       row.Operation,
			BucketStart:     fromUnixMillis(row.BucketStartMS),
			SpanCount:       row.SpanCount,
			ErrorCount:      row.ErrorCount,
			TotalDurationMS: row.TotalDurationMS,
			MaxDurationMS:   row.MaxDurationMS,
		}
		if item.SpanCount > 0 {
			item.AverageDurationM = float64(item.TotalDurationMS) / float64(item.SpanCount)
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *Repository) countFromClickHouse(ctx context.Context, query string) (int64, error) {
	rows, err := platformclickhouse.Select[countRow](ctx, r.ch, query)
	if err != nil || len(rows) == 0 {
		return 0, err
	}
	return rows[0].Count, nil
}

func baseTelemetryClauses(filters observabilityrequests.Filters, timeColumn string) []string {
	clauses := make([]string, 0)
	if filters.ServiceID != "" {
		clauses = append(clauses, fmt.Sprintf("service_id = '%s'", escapeLiteral(filters.ServiceID)))
	}
	// Allow human-readable service name search (case-insensitive substring match)
	if filters.ServiceName != "" {
		clauses = append(clauses, fmt.Sprintf("lower(service_name) LIKE lower('%%%s%%')", escapeLiteral(filters.ServiceName)))
	}
	if filters.Environment != "" {
		clauses = append(clauses, fmt.Sprintf("environment = '%s'", escapeLiteral(filters.Environment)))
	}
	if !filters.StartTime.IsZero() {
		clauses = append(clauses, fmt.Sprintf("%s >= toDateTime64('%s', 3, 'UTC')", timeColumn, filters.StartTime.UTC().Format("2006-01-02 15:04:05.000")))
	}
	if !filters.EndTime.IsZero() {
		clauses = append(clauses, fmt.Sprintf("%s <= toDateTime64('%s', 3, 'UTC')", timeColumn, filters.EndTime.UTC().Format("2006-01-02 15:04:05.000")))
	}
	return clauses
}

func logFilterClause(filters observabilityrequests.Filters) string {
	clauses := baseTelemetryClauses(filters, "occurred_at")
	if filters.TraceID != "" {
		clauses = append(clauses, fmt.Sprintf("trace_id = '%s'", escapeLiteral(filters.TraceID)))
	}
	if filters.Severity != "" {
		clauses = append(clauses, fmt.Sprintf("severity = '%s'", escapeLiteral(filters.Severity)))
	}
	if strings.TrimSpace(filters.Search) != "" {
		clauses = append(clauses, fmt.Sprintf("positionCaseInsensitive(message, '%s') > 0", escapeLiteral(filters.Search)))
	}
	return joinClauses(clauses)
}

func metricFilterClause(filters observabilityrequests.Filters) string {
	clauses := baseTelemetryClauses(filters, "occurred_at")
	if filters.MetricName != "" {
		// Partial case-insensitive search — allows typing "uptime" to match "analytics_uptime_seconds"
		clauses = append(clauses, fmt.Sprintf("lower(metric_name) LIKE lower('%%%s%%')", escapeLiteral(filters.MetricName)))
	}
	return joinClauses(clauses)
}

func traceFilterClause(filters observabilityrequests.Filters) string {
	clauses := baseTelemetryClauses(filters, "occurred_at")
	if filters.TraceID != "" {
		clauses = append(clauses, fmt.Sprintf("trace_id = '%s'", escapeLiteral(filters.TraceID)))
	}
	return joinClauses(clauses)
}

func logRollupFilterClause(filters observabilityrequests.Filters) string {
	clauses := baseTelemetryClauses(filters, "bucket_start")
	if filters.Severity != "" {
		clauses = append(clauses, fmt.Sprintf("severity = '%s'", escapeLiteral(filters.Severity)))
	}
	return joinClauses(clauses)
}

func metricRollupFilterClause(filters observabilityrequests.Filters) string {
	clauses := baseTelemetryClauses(filters, "bucket_start")
	if filters.MetricName != "" {
		clauses = append(clauses, fmt.Sprintf("lower(metric_name) LIKE lower('%%%s%%')", escapeLiteral(filters.MetricName)))
	}
	return joinClauses(clauses)
}

func traceRollupFilterClause(filters observabilityrequests.Filters) string {
	clauses := baseTelemetryClauses(filters, "bucket_start")
	return joinClauses(clauses)
}

func joinClauses(clauses []string) string {
	if len(clauses) == 0 {
		return ""
	}
	return " AND " + strings.Join(clauses, " AND ")
}

func escapeLiteral(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\''`)
	return replacer.Replace(value)
}

// ─── Service Topology ─────────────────────────────────────────────────────────

type TopologyEdge struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	CallCount int64  `json:"call_count"`
}

type TopologyNode struct {
	ServiceName  string  `json:"service_name"`
	TotalCalls   int64   `json:"total_calls"`
	ErrorRate    float64 `json:"error_rate"`
	AvgLatencyMS float64 `json:"avg_latency_ms"`
}

type ServiceTopology struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}

func (r *Repository) serviceTopologyCH(ctx context.Context, tenantID string, lookbackMinutes int) (ServiceTopology, error) {
	if lookbackMinutes <= 0 {
		lookbackMinutes = 60
	}

	// Step 1: Get all service nodes with their health stats
	nodeQuery := fmt.Sprintf(`
		SELECT
			service_name,
			count()                                            AS total_calls,
			countIf(lower(status) NOT IN ('ok',''))            AS error_count,
			avg(duration_ms)                                   AS avg_latency_ms
		FROM pulselens.trace_spans
		WHERE tenant_id = '%s'
		  AND occurred_at >= now() - INTERVAL %d MINUTE
		GROUP BY service_name
		ORDER BY total_calls DESC
		LIMIT 50
	`, escapeLiteral(tenantID), lookbackMinutes)

	type nodeRow struct {
		ServiceName  string  `json:"service_name"`
		TotalCalls   int64   `json:"total_calls"`
		ErrorCount   int64   `json:"error_count"`
		AvgLatencyMS float64 `json:"avg_latency_ms"`
	}
	nodeRows, err := platformclickhouse.Select[nodeRow](ctx, r.ch, nodeQuery)
	if err != nil {
		return ServiceTopology{}, err
	}
	nodes := make([]TopologyNode, 0, len(nodeRows))
	for _, row := range nodeRows {
		var errRate float64
		if row.TotalCalls > 0 {
			errRate = float64(row.ErrorCount) / float64(row.TotalCalls) * 100
		}
		nodes = append(nodes, TopologyNode{
			ServiceName:  row.ServiceName,
			TotalCalls:   row.TotalCalls,
			ErrorRate:    errRate,
			AvgLatencyMS: row.AvgLatencyMS,
		})
	}

	// Step 2: Find edges — services that share a trace_id (caller→callee inferred by start_time order)
	// We self-join trace_spans on trace_id where the two spans have different service names
	edgeQuery := fmt.Sprintf(`
		SELECT
			a.service_name AS source,
			b.service_name AS target,
			count()        AS call_count
		FROM pulselens.trace_spans a
		JOIN pulselens.trace_spans b
		  ON a.trace_id = b.trace_id
		 AND a.service_name < b.service_name
		WHERE a.tenant_id = '%s'
		  AND a.occurred_at >= now() - INTERVAL %d MINUTE
		  AND b.tenant_id = '%s'
		  AND b.occurred_at >= now() - INTERVAL %d MINUTE
		  AND a.service_name != b.service_name
		GROUP BY source, target
		ORDER BY call_count DESC
		LIMIT 200
	`, escapeLiteral(tenantID), lookbackMinutes, escapeLiteral(tenantID), lookbackMinutes)

	type edgeRow struct {
		Source    string `json:"source"`
		Target    string `json:"target"`
		CallCount int64  `json:"call_count"`
	}
	edgeRows, err := platformclickhouse.Select[edgeRow](ctx, r.ch, edgeQuery)
	if err != nil {
		// Edges are optional — return nodes only if edge query fails
		return ServiceTopology{Nodes: nodes, Edges: []TopologyEdge{}}, nil
	}
	edges := make([]TopologyEdge, 0, len(edgeRows))
	for _, row := range edgeRows {
		edges = append(edges, TopologyEdge{
			Source:    row.Source,
			Target:    row.Target,
			CallCount: row.CallCount,
		})
	}

	return ServiceTopology{Nodes: nodes, Edges: edges}, nil
}

func fromUnixMillis(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(value).UTC()
}
