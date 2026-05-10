package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	alertmodels "github.com/omniful/pulselens-alerting-service/internal/alerts/models"
	serviceclickhouse "github.com/omniful/pulselens-alerting-service/pkg/clickhouse"
	platformclickhouse "github.com/omniful/pulselens-platform/clickhouse"
	"github.com/omniful/pulselens-platform/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	db *gorm.DB
}

type chCountRow struct {
	Value float64 `json:"value"`
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateRule(ctx context.Context, row *alertmodels.AlertRule) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) CreatePolicy(ctx context.Context, row *alertmodels.AlertPolicy) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) UpdateRule(ctx context.Context, row *alertmodels.AlertRule) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *Repository) UpdatePolicy(ctx context.Context, row *alertmodels.AlertPolicy) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *Repository) GetRule(ctx context.Context, tenantID, ruleID string) (alertmodels.AlertRule, error) {
	var row alertmodels.AlertRule
	err := r.db.WithContext(ctx).Where("tenant_id = ? and id = ?", tenantID, ruleID).First(&row).Error
	return row, err
}

func (r *Repository) GetPolicy(ctx context.Context, tenantID, policyID string) (alertmodels.AlertPolicy, error) {
	var row alertmodels.AlertPolicy
	err := r.db.WithContext(ctx).Where("tenant_id = ? and id = ?", tenantID, policyID).First(&row).Error
	return row, err
}

func (r *Repository) ListRules(ctx context.Context, tenantID string) ([]alertmodels.AlertRule, error) {
	rows := make([]alertmodels.AlertRule, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) ListPolicies(ctx context.Context, tenantID string) ([]alertmodels.AlertPolicy, error) {
	rows := make([]alertmodels.AlertPolicy, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) FindDefaultPolicy(ctx context.Context, tenantID string) (alertmodels.AlertPolicy, error) {
	var row alertmodels.AlertPolicy
	err := r.db.WithContext(ctx).Where("tenant_id = ? and active = ?", tenantID, true).Order("created_at asc").First(&row).Error
	return row, err
}

func (r *Repository) ListActiveRules(ctx context.Context) ([]alertmodels.AlertRule, error) {
	rows := make([]alertmodels.AlertRule, 0)
	err := r.db.WithContext(ctx).Where("active = ?", true).Find(&rows).Error
	return rows, err
}

func (r *Repository) ListIncidents(ctx context.Context, tenantID string, status string) ([]alertmodels.Incident, error) {
	rows := make([]alertmodels.Incident, 0)
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("triggered_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) GetOpenIncident(ctx context.Context, ruleID string) (alertmodels.Incident, error) {
	var row alertmodels.Incident
	err := r.db.WithContext(ctx).
		Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
		Where("alert_rule_id = ? and status in ?", ruleID, []string{"open", "acknowledged"}).
		First(&row).Error
	return row, err
}

func (r *Repository) CreateIncident(ctx context.Context, row *alertmodels.Incident) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) SaveIncident(ctx context.Context, row *alertmodels.Incident) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *Repository) GetIncident(ctx context.Context, tenantID, incidentID string) (alertmodels.Incident, error) {
	var row alertmodels.Incident
	err := r.db.WithContext(ctx).Where("tenant_id = ? and id = ?", tenantID, incidentID).First(&row).Error
	return row, err
}

func (r *Repository) CreateNotificationChannel(ctx context.Context, row *alertmodels.NotificationChannel) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) ListNotificationChannels(ctx context.Context, tenantID string) ([]alertmodels.NotificationChannel, error) {
	rows := make([]alertmodels.NotificationChannel, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) ListActiveNotificationChannels(ctx context.Context, tenantID string) ([]alertmodels.NotificationChannel, error) {
	rows := make([]alertmodels.NotificationChannel, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ? and active = ?", tenantID, true).Find(&rows).Error
	return rows, err
}

func (r *Repository) CreateNotificationDelivery(ctx context.Context, row *alertmodels.NotificationDelivery) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) UpdateNotificationDelivery(ctx context.Context, row *alertmodels.NotificationDelivery) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *Repository) ListNotificationDeliveries(ctx context.Context, tenantID string) ([]alertmodels.NotificationDelivery, error) {
	rows := make([]alertmodels.NotificationDelivery, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) CreateIncidentComment(ctx context.Context, row *alertmodels.IncidentComment) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) ListIncidentComments(ctx context.Context, tenantID, incidentID string) ([]alertmodels.IncidentComment, error) {
	rows := make([]alertmodels.IncidentComment, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ? and incident_id = ?", tenantID, incidentID).Order("created_at asc").Find(&rows).Error
	return rows, err
}

func (r *Repository) MarkRuleEvaluated(ctx context.Context, ruleID string, evaluatedAt time.Time, triggeredAt *time.Time) error {
	updates := map[string]interface{}{
		"last_evaluation_at": evaluatedAt,
	}
	if triggeredAt != nil {
		updates["last_triggered_at"] = *triggeredAt
	}
	return r.db.WithContext(ctx).Model(&alertmodels.AlertRule{}).Where("id = ?", ruleID).Updates(updates).Error
}

func (r *Repository) ListEscalationCandidates(ctx context.Context, limit int) ([]alertmodels.Incident, error) {
	rows := make([]alertmodels.Incident, 0)
	query := r.db.WithContext(ctx).
		Where("status = ? and next_escalation_at is not null and next_escalation_at <= ?", "open", time.Now().UTC()).
		Order("next_escalation_at asc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&rows).Error
	return rows, err
}

func (r *Repository) CountLogEvents(ctx context.Context, tenantID string, serviceID string, severity string, since time.Time) (float64, error) {
	if value, ok := r.countLogEventsCH(ctx, tenantID, serviceID, severity, since); ok {
		return value, nil
	}
	var count int64
	query := r.db.WithContext(ctx).Table("log_events").Where("tenant_id = ? and occurred_at >= ?", tenantID, since)
	if serviceID != "" {
		query = query.Where("service_id = ?", serviceID)
	}
	if severity != "" {
		query = query.Where("lower(severity) = lower(?)", severity)
	}
	err := query.Count(&count).Error
	return float64(count), err
}

func (r *Repository) CountTraceErrors(ctx context.Context, tenantID string, serviceID string, since time.Time) (float64, error) {
	if value, ok := r.countTraceErrorsCH(ctx, tenantID, serviceID, since); ok {
		return value, nil
	}
	var count int64
	query := r.db.WithContext(ctx).Table("trace_spans").
		Where("tenant_id = ? and occurred_at >= ? and lower(status) <> 'ok'", tenantID, since)
	if serviceID != "" {
		query = query.Where("service_id = ?", serviceID)
	}
	err := query.Count(&count).Error
	return float64(count), err
}

func (r *Repository) AggregateMetric(ctx context.Context, tenantID string, serviceID string, metricName string, aggregation string, since time.Time) (float64, error) {
	if value, ok := r.aggregateMetricCH(ctx, tenantID, serviceID, metricName, aggregation, since); ok {
		return value, nil
	}
	type result struct {
		Value float64 `gorm:"column:value"`
	}
	var row result
	aggregateExpr := "avg(value)"
	switch aggregation {
	case "max":
		aggregateExpr = "max(value)"
	case "sum":
		aggregateExpr = "sum(value)"
	case "count":
		aggregateExpr = "count(*)"
	}

	query := r.db.WithContext(ctx).
		Table("metric_points").
		Select(aggregateExpr+" as value").
		Where("tenant_id = ? and occurred_at >= ?", tenantID, since)
	if serviceID != "" {
		query = query.Where("service_id = ?", serviceID)
	}
	if metricName != "" {
		query = query.Where("metric_name = ?", metricName)
	}
	err := query.Scan(&row).Error
	return row.Value, err
}

func (r *Repository) countLogEventsCH(ctx context.Context, tenantID string, serviceID string, severity string, since time.Time) (float64, bool) {
	if serviceclickhouse.Get() == nil || !serviceclickhouse.Get().Enabled() {
		return 0, false
	}
	query := fmt.Sprintf(`
		SELECT toFloat64(coalesce(sum(event_count), 0)) AS value
		FROM log_severity_rollup_minutes
		WHERE tenant_id = '%s' AND bucket_start >= toDateTime64('%s', 3, 'UTC')%s%s
	`, escapeLiteral(tenantID), since.UTC().Truncate(time.Minute).Format("2006-01-02 15:04:05.000"), serviceClause(serviceID), severityClause(severity))
	rows, err := platformclickhouse.Select[chCountRow](ctx, serviceclickhouse.Get(), query)
	if err != nil || len(rows) == 0 {
		if err != nil {
			logging.Errorf("countLogEventsCH failed: %v", err)
		}
		return 0, false
	}
	return rows[0].Value, true
}

func (r *Repository) countTraceErrorsCH(ctx context.Context, tenantID string, serviceID string, since time.Time) (float64, bool) {
	if serviceclickhouse.Get() == nil || !serviceclickhouse.Get().Enabled() {
		return 0, false
	}
	query := fmt.Sprintf(`
		SELECT toFloat64(coalesce(sum(error_count), 0)) AS value
		FROM trace_latency_rollup_minutes
		WHERE tenant_id = '%s' AND bucket_start >= toDateTime64('%s', 3, 'UTC')%s
	`, escapeLiteral(tenantID), since.UTC().Truncate(time.Minute).Format("2006-01-02 15:04:05.000"), serviceClause(serviceID))
	rows, err := platformclickhouse.Select[chCountRow](ctx, serviceclickhouse.Get(), query)
	if err != nil || len(rows) == 0 {
		if err != nil {
			logging.Errorf("countTraceErrorsCH failed: %v", err)
		}
		return 0, false
	}
	return rows[0].Value, true
}

func (r *Repository) aggregateMetricCH(ctx context.Context, tenantID string, serviceID string, metricName string, aggregation string, since time.Time) (float64, bool) {
	if serviceclickhouse.Get() == nil || !serviceclickhouse.Get().Enabled() {
		return 0, false
	}
	selectExpr := "toFloat64(coalesce(sum(sum_value) / nullIf(sum(sample_count), 0), 0)) AS value"
	switch strings.ToLower(strings.TrimSpace(aggregation)) {
	case "max":
		selectExpr = "toFloat64(coalesce(max(max_value), 0)) AS value"
	case "sum":
		selectExpr = "toFloat64(coalesce(sum(sum_value), 0)) AS value"
	case "count":
		selectExpr = "toFloat64(coalesce(sum(sample_count), 0)) AS value"
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM metric_rollup_minutes
		WHERE tenant_id = '%s' AND bucket_start >= toDateTime64('%s', 3, 'UTC')%s%s
	`, selectExpr, escapeLiteral(tenantID), since.UTC().Truncate(time.Minute).Format("2006-01-02 15:04:05.000"), serviceClause(serviceID), metricClause(metricName))
	rows, err := platformclickhouse.Select[chCountRow](ctx, serviceclickhouse.Get(), query)
	if err != nil || len(rows) == 0 {
		if err != nil {
			logging.Errorf("aggregateMetricCH failed: %v", err)
		}
		return 0, false
	}
	return rows[0].Value, true
}

func serviceClause(serviceID string) string {
	if strings.TrimSpace(serviceID) == "" {
		return ""
	}
	return " AND service_id = '" + escapeLiteral(serviceID) + "'"
}

func severityClause(severity string) string {
	if strings.TrimSpace(severity) == "" {
		return ""
	}
	return " AND severity = '" + escapeLiteral(strings.ToLower(strings.TrimSpace(severity))) + "'"
}

func metricClause(metricName string) string {
	if strings.TrimSpace(metricName) == "" {
		return ""
	}
	return " AND metric_name = '" + escapeLiteral(metricName) + "'"
}

func escapeLiteral(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return replacer.Replace(value)
}

func marshalPayload(value interface{}) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}
