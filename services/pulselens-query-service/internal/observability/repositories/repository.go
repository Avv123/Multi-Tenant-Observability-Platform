package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	commonauth "github.com/Avv123/pulselens-common/auth"
	platformclickhouse "github.com/Avv123/pulselens-platform/clickhouse"
	"github.com/Avv123/pulselens-platform/idgen"
	observabilitymodels "github.com/Avv123/pulselens-query-service/internal/observability/models"
	observabilityrequests "github.com/Avv123/pulselens-query-service/internal/observability/requests"
	observabilityresponses "github.com/Avv123/pulselens-query-service/internal/observability/responses"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
	ch *platformclickhouse.Client
}

func NewRepository(db *gorm.DB, ch *platformclickhouse.Client) *Repository {
	return &Repository{db: db, ch: ch}
}

func (r *Repository) CountLogs(ctx context.Context, tenantID string) (int64, error) {
	if !r.useClickHouse() {
		return 0, r.telemetryUnavailable()
	}
	return r.countLogsCH(ctx, tenantID)
}

func (r *Repository) CountMetrics(ctx context.Context, tenantID string) (int64, error) {
	if !r.useClickHouse() {
		return 0, r.telemetryUnavailable()
	}
	return r.countMetricsCH(ctx, tenantID)
}

func (r *Repository) CountTraceSpans(ctx context.Context, tenantID string) (int64, error) {
	if !r.useClickHouse() {
		return 0, r.telemetryUnavailable()
	}
	return r.countTraceSpansCH(ctx, tenantID)
}

func (r *Repository) CountServices(ctx context.Context, tenantID string) (int64, error) {
	if !r.useClickHouse() {
		return 0, r.telemetryUnavailable()
	}
	return r.countServicesCH(ctx, tenantID)
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

func (r *Repository) ListLogs(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.LogRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listLogsCH(ctx, tenantID, filters)
}

func (r *Repository) ListMetrics(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.MetricRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listMetricsCH(ctx, tenantID, filters)
}

func (r *Repository) ListTraces(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.TraceRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listTracesCH(ctx, tenantID, filters)
}

func (r *Repository) TraceDetail(ctx context.Context, tenantID string, traceID string) ([]observabilityresponses.TraceSpanRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.traceDetailCH(ctx, tenantID, traceID)
}

func (r *Repository) ListServiceHealth(ctx context.Context, tenantID string, limit int) ([]observabilityresponses.ServiceHealthRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listServiceHealthCH(ctx, tenantID, limit)
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

func (r *Repository) CreateDashboard(ctx context.Context, claims *commonauth.Claims, request observabilityrequests.UpdateDashboardRequest) (observabilitymodels.Dashboard, error) {
	layout, err := json.Marshal(request.Layout)
	if err != nil {
		return observabilitymodels.Dashboard{}, err
	}
	widgets, err := json.Marshal(request.Widgets)
	if err != nil {
		return observabilitymodels.Dashboard{}, err
	}
	row := observabilitymodels.Dashboard{
		ID:               idgen.New("dash"),
		TenantID:         claims.TenantID,
		Name:             request.Name,
		Description:      request.Description,
		DefaultTimeRange: strings.TrimSpace(request.DefaultTimeRange),
		Layout:           string(layout),
		Widgets:          string(widgets),
		CreatedBy:        claims.UserID,
	}
	if row.DefaultTimeRange == "" {
		row.DefaultTimeRange = "120m"
	}
	err = r.db.WithContext(ctx).Create(&row).Error
	return row, err
}

func (r *Repository) ListDashboards(ctx context.Context, tenantID string) ([]observabilitymodels.Dashboard, error) {
	rows := make([]observabilitymodels.Dashboard, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) GetDashboard(ctx context.Context, tenantID, dashboardID string) (observabilitymodels.Dashboard, error) {
	row := observabilitymodels.Dashboard{}
	err := r.db.WithContext(ctx).Where("tenant_id = ? and id = ?", tenantID, dashboardID).First(&row).Error
	return row, err
}

func (r *Repository) UpdateDashboard(ctx context.Context, claims *commonauth.Claims, dashboardID string, request observabilityrequests.UpdateDashboardRequest) (observabilitymodels.Dashboard, error) {
	row, err := r.GetDashboard(ctx, claims.TenantID, dashboardID)
	if err != nil {
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
	row.DefaultTimeRange = strings.TrimSpace(request.DefaultTimeRange)
	if row.DefaultTimeRange == "" {
		row.DefaultTimeRange = "120m"
	}
	row.Layout = string(layout)
	row.Widgets = string(widgets)
	err = r.db.WithContext(ctx).Save(&row).Error
	return row, err
}

func (r *Repository) ReplaceDashboard(ctx context.Context, row *observabilitymodels.Dashboard) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *Repository) ListLogSeverityRollups(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.LogSeverityRollupRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listLogSeverityRollupsCH(ctx, tenantID, filters)
}

func (r *Repository) ListMetricSeries(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.MetricSeriesRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listMetricSeriesCH(ctx, tenantID, filters)
}

func (r *Repository) ListTraceLatencyRollups(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.TraceLatencyRollupRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listTraceLatencyRollupsCH(ctx, tenantID, filters)
}

func (r *Repository) ListTransactions(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.TransactionRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listTransactionsCH(ctx, tenantID, filters)
}

func (r *Repository) ListErrorGroups(ctx context.Context, tenantID string, filters observabilityrequests.Filters) ([]observabilityresponses.ErrorGroupRow, error) {
	if !r.useClickHouse() {
		return nil, r.telemetryUnavailable()
	}
	return r.listErrorGroupsCH(ctx, tenantID, filters)
}

func (r *Repository) GetServiceMap(ctx context.Context, tenantID string, lookbackMinutes int) (ServiceTopology, error) {
	if !r.useClickHouse() {
		return ServiceTopology{}, r.telemetryUnavailable()
	}
	return r.serviceTopologyCH(ctx, tenantID, lookbackMinutes)
}

func (r *Repository) telemetryUnavailable() error {
	return errors.New("clickhouse telemetry store unavailable")
}

func (r *Repository) useClickHouse() bool {
	return r.ch != nil && r.ch.Enabled()
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
