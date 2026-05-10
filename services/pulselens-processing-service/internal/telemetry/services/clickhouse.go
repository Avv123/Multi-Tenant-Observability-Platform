package services

import (
	"context"
	"encoding/json"
	"time"

	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	platformclickhouse "github.com/omniful/pulselens-platform/clickhouse"
	serviceclickhouse "github.com/omniful/pulselens-processing-service/pkg/clickhouse"
)

func (s *Service) persistToClickHouse(ctx context.Context, envelope pulsetelemetry.Envelope, payloadBytes []byte) error {
	client := serviceclickhouse.Get()
	if client == nil || !client.Enabled() {
		return nil
	}

	row := map[string]any{
		"event_id":     envelope.EventID,
		"tenant_id":    envelope.TenantID,
		"service_id":   envelope.ServiceID,
		"service_name": envelope.ServiceName,
		"environment":  envelope.Environment,
		"shard_id":     envelope.ShardID,
		"payload":      string(payloadBytes),
		"occurred_at":  clickhouseTime(envelope.OccurredAt),
		"received_at":  clickhouseTime(envelope.ReceivedAt),
		"created_at":   clickhouseTime(time.Now().UTC()),
	}

	table := "custom_events"
	switch envelope.EventType {
	case pulsetelemetry.EventTypeMetric:
		table = "metric_points"
		row["metric_name"] = stringValue(envelope.Payload["metric_name"])
		row["value"] = floatValue(envelope.Payload["value"])
	case pulsetelemetry.EventTypeTrace:
		table = "trace_spans"
		row["trace_id"] = envelope.TraceID
		row["span_id"] = stringValue(envelope.Payload["span_id"])
		row["parent_span_id"] = stringValue(envelope.Payload["parent_span_id"])
		row["operation"] = stringValue(envelope.Payload["operation"])
		row["status"] = stringValue(envelope.Payload["status"])
		row["duration_ms"] = durationFromPayload(envelope.Payload)
	case pulsetelemetry.EventTypeCustom:
	default:
		table = "log_events"
		row["trace_id"] = envelope.TraceID
		row["severity"] = envelope.Severity
		row["message"] = stringValue(envelope.Payload["message"])
	}

	payload, err := json.Marshal(row)
	if err != nil {
		return err
	}
	return client.InsertJSONEachRow(ctx, table, [][]byte{payload})
}

func clickhouseTime(value time.Time) string {
	return value.UTC().Format("2006-01-02 15:04:05.000")
}

func selectRows[T any](ctx context.Context, query string) ([]T, error) {
	return platformclickhouse.Select[T](ctx, serviceclickhouse.Get(), query)
}

func (s *Service) cleanupClickHouseTenant(ctx context.Context, tenantID string, cutoff time.Time) error {
	client := serviceclickhouse.Get()
	if client == nil || !client.Enabled() {
		return nil
	}

	timestamp := clickhouseTime(cutoff)
	queries := []string{
		"ALTER TABLE log_events DELETE WHERE tenant_id = '" + tenantID + "' AND occurred_at < toDateTime64('" + timestamp + "', 3, 'UTC')",
		"ALTER TABLE metric_points DELETE WHERE tenant_id = '" + tenantID + "' AND occurred_at < toDateTime64('" + timestamp + "', 3, 'UTC')",
		"ALTER TABLE trace_spans DELETE WHERE tenant_id = '" + tenantID + "' AND occurred_at < toDateTime64('" + timestamp + "', 3, 'UTC')",
		"ALTER TABLE custom_events DELETE WHERE tenant_id = '" + tenantID + "' AND occurred_at < toDateTime64('" + timestamp + "', 3, 'UTC')",
		"ALTER TABLE telemetry_rollup_minutes DELETE WHERE tenant_id = '" + tenantID + "' AND bucket_start < toDateTime64('" + timestamp + "', 3, 'UTC')",
		"ALTER TABLE service_health_rollup_minutes DELETE WHERE tenant_id = '" + tenantID + "' AND bucket_start < toDateTime64('" + timestamp + "', 3, 'UTC')",
		"ALTER TABLE log_severity_rollup_minutes DELETE WHERE tenant_id = '" + tenantID + "' AND bucket_start < toDateTime64('" + timestamp + "', 3, 'UTC')",
		"ALTER TABLE metric_rollup_minutes DELETE WHERE tenant_id = '" + tenantID + "' AND bucket_start < toDateTime64('" + timestamp + "', 3, 'UTC')",
		"ALTER TABLE trace_latency_rollup_minutes DELETE WHERE tenant_id = '" + tenantID + "' AND bucket_start < toDateTime64('" + timestamp + "', 3, 'UTC')",
	}
	for _, query := range queries {
		if err := client.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
