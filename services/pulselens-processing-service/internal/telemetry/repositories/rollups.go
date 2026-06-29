package repositories

import (
	"context"
	"strings"
	"time"

	pulsetelemetry "github.com/Avv123/pulselens-common/telemetry"
	"github.com/Avv123/pulselens-platform/idgen"
	"github.com/Avv123/pulselens-processing-service/internal/telemetry/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) IncrementRollups(ctx context.Context, envelope pulsetelemetry.Envelope) error {
	bucketStart := envelope.OccurredAt.UTC().Truncate(time.Minute)
	occurredAt := envelope.OccurredAt.UTC()

	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"},
			{Name: "service_id"},
			{Name: "environment"},
			{Name: "event_type"},
			{Name: "bucket_start"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"event_count":   gorm.Expr("telemetry_rollup_minutes.event_count + 1"),
			"last_event_at": gorm.Expr("GREATEST(telemetry_rollup_minutes.last_event_at, EXCLUDED.last_event_at)"),
			"updated_at":    time.Now().UTC(),
		}),
	}).Create(&models.TelemetryRollupMinute{
		ID:          idgen.New("trm"),
		TenantID:    envelope.TenantID,
		ServiceID:   envelope.ServiceID,
		Environment: envelope.Environment,
		EventType:   string(envelope.EventType),
		BucketStart: bucketStart,
		EventCount:  1,
		LastEventAt: occurredAt,
		UpdatedAt:   time.Now().UTC(),
	}).Error; err != nil {
		return err
	}

	errorCount, criticalCount := severityCounts(envelope.Severity)
	latestMetricAt := time.Time{}
	latestTraceAt := time.Time{}
	if envelope.EventType == pulsetelemetry.EventTypeMetric {
		latestMetricAt = occurredAt
	}
	if envelope.EventType == pulsetelemetry.EventTypeTrace {
		latestTraceAt = occurredAt
	}

	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"},
			{Name: "service_id"},
			{Name: "environment"},
			{Name: "bucket_start"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"service_name":       gorm.Expr("EXCLUDED.service_name"),
			"event_count":        gorm.Expr("service_health_rollup_minutes.event_count + 1"),
			"error_log_count":    gorm.Expr("service_health_rollup_minutes.error_log_count + ?", errorCount),
			"critical_log_count": gorm.Expr("service_health_rollup_minutes.critical_log_count + ?", criticalCount),
			"last_event_at":      gorm.Expr("GREATEST(service_health_rollup_minutes.last_event_at, EXCLUDED.last_event_at)"),
			"latest_metric_at":   gorm.Expr("GREATEST(service_health_rollup_minutes.latest_metric_at, EXCLUDED.latest_metric_at)"),
			"latest_trace_at":    gorm.Expr("GREATEST(service_health_rollup_minutes.latest_trace_at, EXCLUDED.latest_trace_at)"),
			"updated_at":         time.Now().UTC(),
		}),
	}).Create(&models.ServiceHealthRollupMinute{
		ID:               idgen.New("shrm"),
		TenantID:         envelope.TenantID,
		ServiceID:        envelope.ServiceID,
		ServiceName:      envelope.ServiceName,
		Environment:      envelope.Environment,
		BucketStart:      bucketStart,
		EventCount:       1,
		ErrorLogCount:    errorCount,
		CriticalLogCount: criticalCount,
		LastEventAt:      occurredAt,
		LatestMetricAt:   latestMetricAt,
		LatestTraceAt:    latestTraceAt,
		UpdatedAt:        time.Now().UTC(),
	}).Error; err != nil {
		return err
	}

	if envelope.EventType == pulsetelemetry.EventTypeLog {
		severity := strings.ToLower(strings.TrimSpace(envelope.Severity))
		if severity == "" {
			severity = "info"
		}
		if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "tenant_id"},
				{Name: "service_id"},
				{Name: "environment"},
				{Name: "severity"},
				{Name: "bucket_start"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"service_name": gorm.Expr("EXCLUDED.service_name"),
				"event_count":  gorm.Expr("log_severity_rollup_minutes.event_count + 1"),
				"updated_at":   time.Now().UTC(),
			}),
		}).Create(&models.LogSeverityRollupMinute{
			ID:          idgen.New("lsrm"),
			TenantID:    envelope.TenantID,
			ServiceID:   envelope.ServiceID,
			ServiceName: envelope.ServiceName,
			Environment: envelope.Environment,
			Severity:    severity,
			BucketStart: bucketStart,
			EventCount:  1,
			UpdatedAt:   time.Now().UTC(),
		}).Error; err != nil {
			return err
		}
	}

	if envelope.EventType == pulsetelemetry.EventTypeMetric {
		value := floatValue(envelope.Payload["value"])
		metricName := stringValue(envelope.Payload["metric_name"])
		return r.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "tenant_id"},
				{Name: "service_id"},
				{Name: "environment"},
				{Name: "metric_name"},
				{Name: "bucket_start"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"sample_count":  gorm.Expr("metric_rollup_minutes.sample_count + 1"),
				"sum_value":     gorm.Expr("metric_rollup_minutes.sum_value + EXCLUDED.sum_value"),
				"min_value":     gorm.Expr("LEAST(metric_rollup_minutes.min_value, EXCLUDED.min_value)"),
				"max_value":     gorm.Expr("GREATEST(metric_rollup_minutes.max_value, EXCLUDED.max_value)"),
				"last_value":    gorm.Expr("EXCLUDED.last_value"),
				"last_event_at": gorm.Expr("GREATEST(metric_rollup_minutes.last_event_at, EXCLUDED.last_event_at)"),
				"updated_at":    time.Now().UTC(),
			}),
		}).Create(&models.MetricRollupMinute{
			ID:          idgen.New("mrm"),
			TenantID:    envelope.TenantID,
			ServiceID:   envelope.ServiceID,
			Environment: envelope.Environment,
			MetricName:  metricName,
			BucketStart: bucketStart,
			SampleCount: 1,
			SumValue:    value,
			MinValue:    value,
			MaxValue:    value,
			LastValue:   value,
			LastEventAt: occurredAt,
			UpdatedAt:   time.Now().UTC(),
		}).Error
	}

	if envelope.EventType != pulsetelemetry.EventTypeTrace {
		return nil
	}

	durationMS := durationMilliseconds(envelope.Payload)
	operation := stringValue(envelope.Payload["operation"])
	status := strings.ToLower(strings.TrimSpace(stringValue(envelope.Payload["status"])))
	traceErrorCount := int64(0)
	if status != "" && status != "ok" {
		traceErrorCount = 1
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"},
			{Name: "service_id"},
			{Name: "environment"},
			{Name: "operation"},
			{Name: "bucket_start"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"service_name":      gorm.Expr("EXCLUDED.service_name"),
			"span_count":        gorm.Expr("trace_latency_rollup_minutes.span_count + 1"),
			"error_count":       gorm.Expr("trace_latency_rollup_minutes.error_count + ?", traceErrorCount),
			"total_duration_ms": gorm.Expr("trace_latency_rollup_minutes.total_duration_ms + ?", durationMS),
			"max_duration_ms":   gorm.Expr("GREATEST(trace_latency_rollup_minutes.max_duration_ms, EXCLUDED.max_duration_ms)"),
			"last_event_at":     gorm.Expr("GREATEST(trace_latency_rollup_minutes.last_event_at, EXCLUDED.last_event_at)"),
			"updated_at":        time.Now().UTC(),
		}),
	}).Create(&models.TraceLatencyRollupMinute{
		ID:              idgen.New("tlrm"),
		TenantID:        envelope.TenantID,
		ServiceID:       envelope.ServiceID,
		ServiceName:     envelope.ServiceName,
		Environment:     envelope.Environment,
		Operation:       operation,
		BucketStart:     bucketStart,
		SpanCount:       1,
		ErrorCount:      traceErrorCount,
		TotalDurationMS: durationMS,
		MaxDurationMS:   durationMS,
		LastEventAt:     occurredAt,
		UpdatedAt:       time.Now().UTC(),
	}).Error
}

func (r *Repository) ListTenantRetentionPolicies(ctx context.Context) ([]models.TenantRetentionPolicy, error) {
	rows := make([]models.TenantRetentionPolicy, 0)
	err := r.db.WithContext(ctx).Table("tenants").Select("id, retention_days").Find(&rows).Error
	return rows, err
}

func severityCounts(severity string) (int64, int64) {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical", "fatal", "panic":
		return 1, 1
	case "error":
		return 1, 0
	default:
		return 0, 0
	}
}

func stringValue(value interface{}) string {
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

func floatValue(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case float32:
		return float64(typed)
	default:
		return 0
	}
}

func durationMilliseconds(payload map[string]interface{}) int64 {
	startTime, startOK := timeValue(payload["start_time"])
	endTime, endOK := timeValue(payload["end_time"])
	if startOK && endOK && !endTime.Before(startTime) {
		return endTime.Sub(startTime).Milliseconds()
	}
	return 0
}

func timeValue(value interface{}) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), true
	case string:
		parsed, err := time.Parse(time.RFC3339, typed)
		if err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}
