package repositories

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	commonauth "github.com/omniful/pulselens-common/auth"
	platformclickhouse "github.com/omniful/pulselens-platform/clickhouse"
	"github.com/omniful/pulselens-platform/idgen"
	"github.com/omniful/pulselens-platform/logging"
	observabilitymodels "github.com/omniful/pulselens-query-service/internal/observability/models"
	observabilityrequests "github.com/omniful/pulselens-query-service/internal/observability/requests"
	observabilityresponses "github.com/omniful/pulselens-query-service/internal/observability/responses"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
	ch *platformclickhouse.Client
}

func NewRepository(db *gorm.DB, ch *platformclickhouse.Client) *Repository {
	return &Repository{db: db, ch: ch}
}

func (r *Repository) CountLogs(ctx context.Context, tenantID string) int64 {
	if r.useClickHouse() {
		if count, err := r.countLogsCH(ctx, tenantID); err == nil {
			return count
		} else {
			logging.Errorf("count logs from clickhouse failed: %v", err)
		}
	}
	if count, ok := r.countTelemetryRollup(ctx, tenantID, "log"); ok {
		return count
	}
	return r.countLogsPG(ctx, tenantID)
}

func (r *Repository) CountMetrics(ctx context.Context, tenantID string) int64 {
	if r.useClickHouse() {
		if count, err := r.countMetricsCH(ctx, tenantID); err == nil {
			return count
		} else {
			logging.Errorf("count metrics from clickhouse failed: %v", err)
		}
	}
	if count, ok := r.countTelemetryRollup(ctx, tenantID, "metric"); ok {
		return count
	}
	return r.countMetricsPG(ctx, tenantID)
}

func (r *Repository) CountTraceSpans(ctx context.Context, tenantID string) int64 {
	if r.useClickHouse() {
		if count, err := r.countTraceSpansCH(ctx, tenantID); err == nil {
			return count
		} else {
			logging.Errorf("count trace spans from clickhouse failed: %v", err)
		}
	}
	if count, ok := r.countTelemetryRollup(ctx, tenantID, "trace"); ok {
		return count
	}
	return r.countTraceSpansPG(ctx, tenantID)
}

func (r *Repository) CountServices(ctx context.Context, tenantID string) int64 {
	if r.useClickHouse() {
		if count, err := r.countServicesCH(ctx, tenantID); err == nil {
			return count
		} else {
			logging.Errorf("count services from clickhouse failed: %v", err)
		}
	}
	if count, ok := r.countServiceRollup(ctx, tenantID); ok {
		return count
	}
	return r.countServicesPG(ctx, tenantID)
}

func (r *Repository) ListUsage(ctx context.Context, tenantID string, limit int) []observabilityresponses.UsageRow {
	rows := make([]observabilityresponses.UsageRow, 0)
	r.db.WithContext(ctx).
		Table("usage_counters").
		Where("tenant_id = ?", tenantID).
		Order("usage_date desc, event_count desc").
		Limit(limit).
		Scan(&rows)
	return rows
}

func (r *Repository) ListLogs(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.LogRow {
	if r.useClickHouse() {
		if rows, err := r.listLogsCH(ctx, tenantID, filters); err == nil {
			return rows
		} else {
			logging.Errorf("list logs from clickhouse failed: %v", err)
		}
	}
	return r.listLogsPG(ctx, tenantID, filters)
}

func (r *Repository) ListMetrics(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.MetricRow {
	if r.useClickHouse() {
		if rows, err := r.listMetricsCH(ctx, tenantID, filters); err == nil {
			return rows
		} else {
			logging.Errorf("list metrics from clickhouse failed: %v", err)
		}
	}
	return r.listMetricsPG(ctx, tenantID, filters)
}

func (r *Repository) ListTraces(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.TraceRow {
	if r.useClickHouse() {
		if rows, err := r.listTracesCH(ctx, tenantID, filters); err == nil {
			return rows
		} else {
			logging.Errorf("list traces from clickhouse failed: %v", err)
		}
	}
	return r.listTracesPG(ctx, tenantID, filters)
}

func (r *Repository) TraceDetail(ctx context.Context, tenantID string, traceID string) []observabilityresponses.TraceSpanRow {
	if r.useClickHouse() {
		if rows, err := r.traceDetailCH(ctx, tenantID, traceID); err == nil {
			return rows
		} else {
			logging.Errorf("trace detail from clickhouse failed: %v", err)
		}
	}
	return r.traceDetailPG(ctx, tenantID, traceID)
}

func (r *Repository) ListServiceHealth(ctx context.Context, tenantID string, limit int) []observabilityresponses.ServiceHealthRow {
	if r.useClickHouse() {
		if rows, err := r.listServiceHealthCH(ctx, tenantID, limit); err == nil {
			return rows
		} else {
			logging.Errorf("service health from clickhouse failed: %v", err)
		}
	}
	if rows, ok := r.listServiceHealthRollups(ctx, tenantID, limit); ok {
		return rows
	}
	return r.listServiceHealthPG(ctx, tenantID, limit)
}

func (r *Repository) CreateSavedQuery(ctx context.Context, claims *commonauth.Claims, request observabilityrequests.CreateSavedQueryRequest) (observabilitymodels.SavedQuery, error) {
	definition, err := json.Marshal(request.Definition)
	if err != nil {
		return observabilitymodels.SavedQuery{}, err
	}
	row := observabilitymodels.SavedQuery{
		ID:         idgen.New("query"),
		TenantID:   claims.TenantID,
		Name:       request.Name,
		QueryType:  request.QueryType,
		Definition: string(definition),
		CreatedBy:  claims.UserID,
	}
	err = r.db.WithContext(ctx).Create(&row).Error
	return row, err
}

func (r *Repository) ListSavedQueries(ctx context.Context, tenantID string) ([]observabilitymodels.SavedQuery, error) {
	rows := make([]observabilitymodels.SavedQuery, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) UpdateSavedQuery(ctx context.Context, claims *commonauth.Claims, queryID string, request observabilityrequests.UpdateSavedQueryRequest) (observabilitymodels.SavedQuery, error) {
	row := observabilitymodels.SavedQuery{}
	if err := r.db.WithContext(ctx).Where("tenant_id = ? and id = ?", claims.TenantID, queryID).First(&row).Error; err != nil {
		return observabilitymodels.SavedQuery{}, err
	}
	definition, err := json.Marshal(request.Definition)
	if err != nil {
		return observabilitymodels.SavedQuery{}, err
	}
	row.Name = request.Name
	row.QueryType = request.QueryType
	row.Definition = string(definition)
	err = r.db.WithContext(ctx).Save(&row).Error
	return row, err
}

func (r *Repository) CreateDashboard(ctx context.Context, claims *commonauth.Claims, request observabilityrequests.CreateDashboardRequest) (observabilitymodels.Dashboard, error) {
	layout, err := json.Marshal(request.Layout)
	if err != nil {
		return observabilitymodels.Dashboard{}, err
	}
	widgets, err := json.Marshal(request.Widgets)
	if err != nil {
		return observabilitymodels.Dashboard{}, err
	}
	row := observabilitymodels.Dashboard{
		ID:          idgen.New("dash"),
		TenantID:    claims.TenantID,
		Name:        request.Name,
		Description: request.Description,
		Layout:      string(layout),
		Widgets:     string(widgets),
		CreatedBy:   claims.UserID,
	}
	err = r.db.WithContext(ctx).Create(&row).Error
	return row, err
}

func (r *Repository) ListDashboards(ctx context.Context, tenantID string) ([]observabilitymodels.Dashboard, error) {
	rows := make([]observabilitymodels.Dashboard, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) UpdateDashboard(ctx context.Context, claims *commonauth.Claims, dashboardID string, request observabilityrequests.UpdateDashboardRequest) (observabilitymodels.Dashboard, error) {
	row := observabilitymodels.Dashboard{}
	if err := r.db.WithContext(ctx).Where("tenant_id = ? and id = ?", claims.TenantID, dashboardID).First(&row).Error; err != nil {
		return observabilitymodels.Dashboard{}, err
	}
	layout, err := json.Marshal(request.Layout)
	if err != nil {
		return observabilitymodels.Dashboard{}, err
	}
	widgets, err := json.Marshal(request.Widgets)
	if err != nil {
		return observabilitymodels.Dashboard{}, err
	}
	row.Name = request.Name
	row.Description = request.Description
	row.Layout = string(layout)
	row.Widgets = string(widgets)
	err = r.db.WithContext(ctx).Save(&row).Error
	return row, err
}

func (r *Repository) useClickHouse() bool {
	return r.ch != nil && r.ch.Enabled()
}

func (r *Repository) countLogsPG(ctx context.Context, tenantID string) int64 {
	var count int64
	r.db.WithContext(ctx).Table("log_events").Where("tenant_id = ?", tenantID).Count(&count)
	return count
}

func (r *Repository) countMetricsPG(ctx context.Context, tenantID string) int64 {
	var count int64
	r.db.WithContext(ctx).Table("metric_points").Where("tenant_id = ?", tenantID).Count(&count)
	return count
}

func (r *Repository) countTraceSpansPG(ctx context.Context, tenantID string) int64 {
	var count int64
	r.db.WithContext(ctx).Table("trace_spans").Where("tenant_id = ?", tenantID).Count(&count)
	return count
}

func (r *Repository) countServicesPG(ctx context.Context, tenantID string) int64 {
	type result struct{ Count int64 }
	var row result
	r.db.WithContext(ctx).Raw(`select count(distinct service_id) as count from log_events where tenant_id = ?`, tenantID).Scan(&row)
	return row.Count
}

func (r *Repository) listLogsPG(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.LogRow {
	query := r.db.WithContext(ctx).Table("log_events").Where("tenant_id = ?", tenantID)
	if filters.ServiceID != "" {
		query = query.Where("service_id = ?", filters.ServiceID)
	}
	if filters.Environment != "" {
		query = query.Where("environment = ?", filters.Environment)
	}
	if filters.Severity != "" {
		query = query.Where("severity = ?", filters.Severity)
	}
	if filters.TraceID != "" {
		query = query.Where("trace_id = ?", filters.TraceID)
	}
	if strings.TrimSpace(filters.Search) != "" {
		query = query.Where("message ILIKE ?", "%"+filters.Search+"%")
	}
	query = timeBounds(query, filters)

	rows := make([]observabilityresponses.LogRow, 0)
	query.Order("occurred_at desc").Limit(filters.Limit).Offset(filters.Offset).Scan(&rows)
	return rows
}

func (r *Repository) listMetricsPG(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.MetricRow {
	query := r.db.WithContext(ctx).Table("metric_points").Where("tenant_id = ?", tenantID)
	if filters.ServiceID != "" {
		query = query.Where("service_id = ?", filters.ServiceID)
	}
	if filters.Environment != "" {
		query = query.Where("environment = ?", filters.Environment)
	}
	if filters.MetricName != "" {
		query = query.Where("metric_name = ?", filters.MetricName)
	}
	query = timeBounds(query, filters)

	rows := make([]observabilityresponses.MetricRow, 0)
	query.Order("occurred_at desc").Limit(filters.Limit).Offset(filters.Offset).Scan(&rows)
	return rows
}

func (r *Repository) listTracesPG(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.TraceRow {
	query := r.db.WithContext(ctx).
		Table("trace_spans").
		Select("trace_id, service_name, environment, count(*) as span_count, min(occurred_at) as first_seen_at, max(occurred_at) as last_seen_at").
		Where("tenant_id = ?", tenantID)
	if filters.ServiceID != "" {
		query = query.Where("service_id = ?", filters.ServiceID)
	}
	if filters.Environment != "" {
		query = query.Where("environment = ?", filters.Environment)
	}
	if filters.TraceID != "" {
		query = query.Where("trace_id = ?", filters.TraceID)
	}
	query = timeBounds(query, filters)

	rows := make([]observabilityresponses.TraceRow, 0)
	query.Group("trace_id, service_name, environment").Order("last_seen_at desc").Limit(filters.Limit).Offset(filters.Offset).Scan(&rows)
	return rows
}

func (r *Repository) traceDetailPG(ctx context.Context, tenantID string, traceID string) []observabilityresponses.TraceSpanRow {
	rows := make([]observabilityresponses.TraceSpanRow, 0)
	r.db.WithContext(ctx).
		Table("trace_spans").
		Where("tenant_id = ? and trace_id = ?", tenantID, traceID).
		Order("occurred_at asc").
		Scan(&rows)
	return rows
}

func (r *Repository) listServiceHealthPG(ctx context.Context, tenantID string, limit int) []observabilityresponses.ServiceHealthRow {
	rows := make([]observabilityresponses.ServiceHealthRow, 0)
	r.db.WithContext(ctx).Raw(`
		select
			le.service_id,
			le.service_name,
			le.environment,
			max(le.occurred_at) as last_event_at,
			count(*) as event_count,
			sum(case when lower(le.severity) = 'error' then 1 else 0 end) as error_log_count,
			sum(case when lower(le.severity) = 'critical' then 1 else 0 end) as critical_log_count,
			coalesce(mp.latest_metric_at, '0001-01-01'::timestamp) as latest_metric_at,
			coalesce(ts.latest_trace_at, '0001-01-01'::timestamp) as latest_trace_at
		from log_events le
		left join (
			select tenant_id, service_id, max(occurred_at) as latest_metric_at
			from metric_points
			where tenant_id = ?
			group by tenant_id, service_id
		) mp on mp.tenant_id = le.tenant_id and mp.service_id = le.service_id
		left join (
			select tenant_id, service_id, max(occurred_at) as latest_trace_at
			from trace_spans
			where tenant_id = ?
			group by tenant_id, service_id
		) ts on ts.tenant_id = le.tenant_id and ts.service_id = le.service_id
		where le.tenant_id = ?
		group by le.service_id, le.service_name, le.environment, mp.latest_metric_at, ts.latest_trace_at
		order by last_event_at desc
		limit ?
	`, tenantID, tenantID, tenantID, limit).Scan(&rows)

	for index := range rows {
		rows[index].HealthStatus = deriveHealth(rows[index])
	}
	return rows
}

func (r *Repository) countTelemetryRollup(ctx context.Context, tenantID string, eventType string) (int64, bool) {
	type result struct {
		Count int64 `gorm:"column:count"`
	}
	var row result
	err := r.db.WithContext(ctx).
		Table("telemetry_rollup_minutes").
		Select("coalesce(sum(event_count), 0) as count").
		Where("tenant_id = ? and event_type = ?", tenantID, eventType).
		Scan(&row).Error
	if err != nil {
		return 0, false
	}
	return row.Count, true
}

func (r *Repository) countServiceRollup(ctx context.Context, tenantID string) (int64, bool) {
	type result struct {
		Count int64 `gorm:"column:count"`
	}
	var row result
	err := r.db.WithContext(ctx).
		Table("service_health_rollup_minutes").
		Select("count(distinct service_id) as count").
		Where("tenant_id = ?", tenantID).
		Scan(&row).Error
	if err != nil {
		return 0, false
	}
	return row.Count, true
}

func (r *Repository) listServiceHealthRollups(ctx context.Context, tenantID string, limit int) ([]observabilityresponses.ServiceHealthRow, bool) {
	rows := make([]observabilityresponses.ServiceHealthRow, 0)
	err := r.db.WithContext(ctx).Raw(`
		select
			service_id,
			max(service_name) as service_name,
			environment,
			max(last_event_at) as last_event_at,
			sum(event_count) as event_count,
			sum(error_log_count) as error_log_count,
			sum(critical_log_count) as critical_log_count,
			max(latest_metric_at) as latest_metric_at,
			max(latest_trace_at) as latest_trace_at
		from service_health_rollup_minutes
		where tenant_id = ?
		group by service_id, environment
		order by last_event_at desc
		limit ?
	`, tenantID, limit).Scan(&rows).Error
	if err != nil {
		return nil, false
	}
	for index := range rows {
		rows[index].HealthStatus = deriveHealth(rows[index])
	}
	return rows, true
}

func timeBounds(query *gorm.DB, filters observabilityrequests.Filters) *gorm.DB {
	if !filters.StartTime.IsZero() {
		query = query.Where("occurred_at >= ?", filters.StartTime)
	}
	if !filters.EndTime.IsZero() {
		query = query.Where("occurred_at <= ?", filters.EndTime)
	}
	return query
}

func deriveHealth(row observabilityresponses.ServiceHealthRow) string {
	switch {
	case row.CriticalCount > 0:
		return "critical"
	case row.ErrorLogCount > 0:
		return "warning"
	default:
		return "healthy"
	}
}

func (r *Repository) ListLogSeverityRollups(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.LogSeverityRollupRow {
	if r.useClickHouse() {
		if rows, err := r.listLogSeverityRollupsCH(ctx, tenantID, filters); err == nil {
			return rows
		} else {
			logging.Errorf("log severity rollups from clickhouse failed: %v", err)
		}
	}
	rows := make([]observabilityresponses.LogSeverityRollupRow, 0)
	query := r.db.WithContext(ctx).Table("log_severity_rollup_minutes").Where("tenant_id = ?", tenantID)
	if filters.ServiceID != "" {
		query = query.Where("service_id = ?", filters.ServiceID)
	}
	if filters.Environment != "" {
		query = query.Where("environment = ?", filters.Environment)
	}
	if filters.Severity != "" {
		query = query.Where("severity = ?", filters.Severity)
	}
	query = bucketBounds(query, filters)
	query.Order("bucket_start desc, service_name asc, severity asc").Limit(filters.Limit).Offset(filters.Offset).Scan(&rows)
	return rows
}

func (r *Repository) ListMetricSeries(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.MetricSeriesRow {
	if r.useClickHouse() {
		if rows, err := r.listMetricSeriesCH(ctx, tenantID, filters); err == nil {
			return rows
		} else {
			logging.Errorf("metric series from clickhouse failed: %v", err)
		}
	}
	rows := make([]observabilityresponses.MetricSeriesRow, 0)
	query := r.db.WithContext(ctx).Table("metric_rollup_minutes").Where("tenant_id = ?", tenantID)
	if filters.ServiceID != "" {
		query = query.Where("service_id = ?", filters.ServiceID)
	}
	if filters.Environment != "" {
		query = query.Where("environment = ?", filters.Environment)
	}
	if filters.MetricName != "" {
		query = query.Where("metric_name = ?", filters.MetricName)
	}
	query = bucketBounds(query, filters)
	query.Order("bucket_start desc, metric_name asc").Limit(filters.Limit).Offset(filters.Offset).Scan(&rows)
	for index := range rows {
		if rows[index].SampleCount > 0 {
			rows[index].AverageValue = rows[index].SumValue / float64(rows[index].SampleCount)
		}
	}
	return rows
}

func (r *Repository) ListTraceLatencyRollups(ctx context.Context, tenantID string, filters observabilityrequests.Filters) []observabilityresponses.TraceLatencyRollupRow {
	if r.useClickHouse() {
		if rows, err := r.listTraceLatencyRollupsCH(ctx, tenantID, filters); err == nil {
			return rows
		} else {
			logging.Errorf("trace latency rollups from clickhouse failed: %v", err)
		}
	}
	rows := make([]observabilityresponses.TraceLatencyRollupRow, 0)
	query := r.db.WithContext(ctx).Table("trace_latency_rollup_minutes").Where("tenant_id = ?", tenantID)
	if filters.ServiceID != "" {
		query = query.Where("service_id = ?", filters.ServiceID)
	}
	if filters.Environment != "" {
		query = query.Where("environment = ?", filters.Environment)
	}
	query = bucketBounds(query, filters)
	query.Order("bucket_start desc, service_name asc, operation asc").Limit(filters.Limit).Offset(filters.Offset).Scan(&rows)
	for index := range rows {
		if rows[index].SpanCount > 0 {
			rows[index].AverageDurationM = float64(rows[index].TotalDurationMS) / float64(rows[index].SpanCount)
		}
	}
	return rows
}

func bucketBounds(query *gorm.DB, filters observabilityrequests.Filters) *gorm.DB {
	if !filters.StartTime.IsZero() {
		query = query.Where("bucket_start >= ?", filters.StartTime.UTC().Truncate(time.Minute))
	}
	if !filters.EndTime.IsZero() {
		query = query.Where("bucket_start <= ?", filters.EndTime.UTC().Truncate(time.Minute))
	}
	return query
}
