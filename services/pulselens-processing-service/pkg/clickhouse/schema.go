package clickhouse

import (
	"context"

	platformclickhouse "github.com/Avv123/pulselens-platform/clickhouse"
)

func EnsureTelemetrySchema(ctx context.Context, client *platformclickhouse.Client) error {
	if client == nil || !client.Enabled() {
		return nil
	}

	queries := []string{
		`CREATE DATABASE IF NOT EXISTS pulselens`,
		`CREATE TABLE IF NOT EXISTS log_events (
			event_id String,
			tenant_id String,
			service_id String,
			service_name String,
			environment String,
			shard_id Int32,
			trace_id String,
			severity String,
			message String,
			payload String,
			occurred_at DateTime64(3, 'UTC'),
			received_at DateTime64(3, 'UTC'),
			created_at DateTime64(3, 'UTC')
		) ENGINE = MergeTree
		PARTITION BY toDate(occurred_at)
		ORDER BY (tenant_id, service_id, occurred_at, event_id)`,
		`CREATE TABLE IF NOT EXISTS metric_points (
			event_id String,
			tenant_id String,
			service_id String,
			service_name String,
			environment String,
			shard_id Int32,
			metric_name String,
			value Float64,
			payload String,
			occurred_at DateTime64(3, 'UTC'),
			received_at DateTime64(3, 'UTC'),
			created_at DateTime64(3, 'UTC')
		) ENGINE = MergeTree
		PARTITION BY toDate(occurred_at)
		ORDER BY (tenant_id, service_id, metric_name, occurred_at, event_id)`,
		`CREATE TABLE IF NOT EXISTS trace_spans (
			event_id String,
			tenant_id String,
			service_id String,
			service_name String,
			environment String,
			shard_id Int32,
			trace_id String,
			span_id String,
			parent_span_id String,
			operation String,
			status String,
			duration_ms Int64,
			payload String,
			occurred_at DateTime64(3, 'UTC'),
			received_at DateTime64(3, 'UTC'),
			created_at DateTime64(3, 'UTC')
		) ENGINE = MergeTree
		PARTITION BY toDate(occurred_at)
		ORDER BY (tenant_id, service_id, trace_id, occurred_at, event_id)`,
		`CREATE TABLE IF NOT EXISTS custom_events (
			event_id String,
			tenant_id String,
			service_id String,
			service_name String,
			environment String,
			shard_id Int32,
			payload String,
			occurred_at DateTime64(3, 'UTC'),
			received_at DateTime64(3, 'UTC'),
			created_at DateTime64(3, 'UTC')
		) ENGINE = MergeTree
		PARTITION BY toDate(occurred_at)
		ORDER BY (tenant_id, service_id, occurred_at, event_id)`,
		`CREATE TABLE IF NOT EXISTS telemetry_rollup_minutes (
			tenant_id String,
			service_id String,
			service_name String,
			environment String,
			event_type String,
			bucket_start DateTime64(3, 'UTC'),
			event_count UInt64,
			last_event_at DateTime64(3, 'UTC')
		) ENGINE = MergeTree
		PARTITION BY toDate(bucket_start)
		ORDER BY (tenant_id, service_id, environment, event_type, bucket_start)`,
		`CREATE TABLE IF NOT EXISTS service_health_rollup_minutes (
			tenant_id String,
			service_id String,
			service_name String,
			environment String,
			bucket_start DateTime64(3, 'UTC'),
			event_count UInt64,
			error_log_count UInt64,
			critical_log_count UInt64,
			last_event_at DateTime64(3, 'UTC'),
			latest_metric_at Nullable(DateTime64(3, 'UTC')),
			latest_trace_at Nullable(DateTime64(3, 'UTC'))
		) ENGINE = MergeTree
		PARTITION BY toDate(bucket_start)
		ORDER BY (tenant_id, service_id, environment, bucket_start)`,
		`CREATE TABLE IF NOT EXISTS log_severity_rollup_minutes (
			tenant_id String,
			service_id String,
			service_name String,
			environment String,
			severity String,
			bucket_start DateTime64(3, 'UTC'),
			event_count UInt64
		) ENGINE = MergeTree
		PARTITION BY toDate(bucket_start)
		ORDER BY (tenant_id, service_id, environment, severity, bucket_start)`,
		`CREATE TABLE IF NOT EXISTS metric_rollup_minutes (
			tenant_id String,
			service_id String,
			environment String,
			metric_name String,
			bucket_start DateTime64(3, 'UTC'),
			sample_count UInt64,
			sum_value Float64,
			min_value Float64,
			max_value Float64,
			last_value Float64,
			last_event_at DateTime64(3, 'UTC')
		) ENGINE = MergeTree
		PARTITION BY toDate(bucket_start)
		ORDER BY (tenant_id, service_id, environment, metric_name, bucket_start)`,
		`CREATE TABLE IF NOT EXISTS trace_latency_rollup_minutes (
			tenant_id String,
			service_id String,
			service_name String,
			environment String,
			operation String,
			bucket_start DateTime64(3, 'UTC'),
			span_count UInt64,
			error_count UInt64,
			total_duration_ms Int64,
			max_duration_ms Int64,
			last_event_at DateTime64(3, 'UTC')
		) ENGINE = MergeTree
		PARTITION BY toDate(bucket_start)
		ORDER BY (tenant_id, service_id, environment, operation, bucket_start)`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_log_telemetry_rollup_minutes
		TO telemetry_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			'log' AS event_type,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS event_count,
			occurred_at AS last_event_at
		FROM log_events`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_metric_telemetry_rollup_minutes
		TO telemetry_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			'metric' AS event_type,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS event_count,
			occurred_at AS last_event_at
		FROM metric_points`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_trace_telemetry_rollup_minutes
		TO telemetry_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			'trace' AS event_type,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS event_count,
			occurred_at AS last_event_at
		FROM trace_spans`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_custom_telemetry_rollup_minutes
		TO telemetry_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			'custom' AS event_type,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS event_count,
			occurred_at AS last_event_at
		FROM custom_events`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_log_service_health_rollup_minutes
		TO service_health_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS event_count,
			toUInt64(lowerUTF8(severity) IN ('error', 'critical', 'fatal', 'panic')) AS error_log_count,
			toUInt64(lowerUTF8(severity) IN ('critical', 'fatal', 'panic')) AS critical_log_count,
			occurred_at AS last_event_at,
			CAST(NULL, 'Nullable(DateTime64(3, ''UTC''))') AS latest_metric_at,
			CAST(NULL, 'Nullable(DateTime64(3, ''UTC''))') AS latest_trace_at
		FROM log_events`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_metric_service_health_rollup_minutes
		TO service_health_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS event_count,
			toUInt64(0) AS error_log_count,
			toUInt64(0) AS critical_log_count,
			occurred_at AS last_event_at,
			occurred_at AS latest_metric_at,
			CAST(NULL, 'Nullable(DateTime64(3, ''UTC''))') AS latest_trace_at
		FROM metric_points`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_trace_service_health_rollup_minutes
		TO service_health_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS event_count,
			toUInt64(0) AS error_log_count,
			toUInt64(0) AS critical_log_count,
			occurred_at AS last_event_at,
			CAST(NULL, 'Nullable(DateTime64(3, ''UTC''))') AS latest_metric_at,
			occurred_at AS latest_trace_at
		FROM trace_spans`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_log_severity_rollup_minutes
		TO log_severity_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			if(lowerUTF8(trim(BOTH ' ' FROM severity)) = '', 'info', lowerUTF8(trim(BOTH ' ' FROM severity))) AS severity,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS event_count
		FROM log_events`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_metric_rollup_minutes
		TO metric_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			environment,
			metric_name,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS sample_count,
			value AS sum_value,
			value AS min_value,
			value AS max_value,
			value AS last_value,
			occurred_at AS last_event_at
		FROM metric_points`,
		`CREATE MATERIALIZED VIEW IF NOT EXISTS mv_trace_latency_rollup_minutes
		TO trace_latency_rollup_minutes AS
		SELECT
			tenant_id,
			service_id,
			service_name,
			environment,
			operation,
			toStartOfMinute(occurred_at) AS bucket_start,
			toUInt64(1) AS span_count,
			toUInt64(lowerUTF8(status) != 'ok' AND lowerUTF8(status) != '') AS error_count,
			duration_ms AS total_duration_ms,
			duration_ms AS max_duration_ms,
			occurred_at AS last_event_at
		FROM trace_spans`,
	}

	for _, query := range queries {
		if err := client.Exec(ctx, query); err != nil {
			return err
		}
	}
	if err := client.Exec(ctx, `ALTER TABLE trace_spans ADD COLUMN IF NOT EXISTS duration_ms Int64`); err != nil {
		return err
	}
	return nil
}
