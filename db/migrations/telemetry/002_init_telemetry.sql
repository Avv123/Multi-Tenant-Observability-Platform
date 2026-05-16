-- Migration 002: telemetry tables
-- Hot-path telemetry storage: logs, metrics, traces, custom events.
-- B10: Explicit migration with correct indexes for tenant-scoped queries.

CREATE TABLE IF NOT EXISTS log_events (
    event_id     TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL,
    service_id   TEXT NOT NULL,
    service_name TEXT NOT NULL,
    environment  TEXT NOT NULL DEFAULT 'production',
    shard_id     INT NOT NULL DEFAULT 0,
    trace_id     TEXT,
    severity     TEXT NOT NULL DEFAULT 'info',
    message      TEXT,
    payload      JSONB,
    occurred_at  TIMESTAMPTZ NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_log_events_tenant_time   ON log_events(tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_log_events_service       ON log_events(tenant_id, service_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_log_events_severity      ON log_events(tenant_id, severity, occurred_at DESC);

CREATE TABLE IF NOT EXISTS metric_points (
    event_id     TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL,
    service_id   TEXT NOT NULL,
    service_name TEXT NOT NULL,
    environment  TEXT NOT NULL DEFAULT 'production',
    shard_id     INT NOT NULL DEFAULT 0,
    metric_name  TEXT NOT NULL,
    value        DOUBLE PRECISION NOT NULL DEFAULT 0,
    payload      JSONB,
    occurred_at  TIMESTAMPTZ NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_metric_points_tenant_metric ON metric_points(tenant_id, metric_name, occurred_at DESC);

CREATE TABLE IF NOT EXISTS trace_spans (
    event_id      TEXT PRIMARY KEY,
    tenant_id     TEXT NOT NULL,
    service_id    TEXT NOT NULL,
    service_name  TEXT NOT NULL,
    environment   TEXT NOT NULL DEFAULT 'production',
    shard_id      INT NOT NULL DEFAULT 0,
    trace_id      TEXT,
    span_id       TEXT,
    parent_span_id TEXT,
    operation     TEXT,
    status        TEXT,
    duration_ms   BIGINT NOT NULL DEFAULT 0,
    payload       JSONB,
    occurred_at   TIMESTAMPTZ NOT NULL,
    received_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_trace_spans_tenant_time  ON trace_spans(tenant_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_trace_spans_trace_id     ON trace_spans(tenant_id, trace_id);

CREATE TABLE IF NOT EXISTS custom_events (
    event_id     TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL,
    service_id   TEXT NOT NULL,
    service_name TEXT NOT NULL,
    environment  TEXT NOT NULL DEFAULT 'production',
    shard_id     INT NOT NULL DEFAULT 0,
    payload      JSONB,
    occurred_at  TIMESTAMPTZ NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dead_letter_events (
    event_id    TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL,
    service_id  TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    reason      TEXT,
    payload     JSONB,
    retry_count INT NOT NULL DEFAULT 0,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS retry_events (
    id             TEXT PRIMARY KEY,
    event_id       TEXT NOT NULL,
    tenant_id      TEXT NOT NULL,
    service_id     TEXT NOT NULL,
    event_type     TEXT NOT NULL,
    payload        JSONB,
    retry_count    INT NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'pending',
    next_attempt_at TIMESTAMPTZ,
    dispatched_at  TIMESTAMPTZ,
    error_message  TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_retry_events_due ON retry_events(status, next_attempt_at) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS usage_counters (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL,
    service_id  TEXT NOT NULL,
    signal_type TEXT NOT NULL,
    date        DATE NOT NULL,
    count       BIGINT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, service_id, signal_type, date)
);

CREATE TABLE IF NOT EXISTS archive_records (
    id             TEXT PRIMARY KEY,
    event_id       TEXT NOT NULL,
    tenant_id      TEXT NOT NULL,
    service_id     TEXT NOT NULL,
    event_type     TEXT NOT NULL,
    payload        JSONB,
    archive_bucket TEXT NOT NULL,
    archive_key    TEXT NOT NULL,
    archive_path   TEXT NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_archive_records_tenant_time ON archive_records(tenant_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS cleanup_runs (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL,
    deleted_logs BIGINT NOT NULL DEFAULT 0,
    deleted_metrics BIGINT NOT NULL DEFAULT 0,
    deleted_traces BIGINT NOT NULL DEFAULT 0,
    deleted_retries BIGINT NOT NULL DEFAULT 0,
    deleted_dlq BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
