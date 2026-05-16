-- Migration 003: alerting and rollup tables
-- B10: Alert rules, incidents, notifications, rollups.

-- Rollup tables
CREATE TABLE IF NOT EXISTS telemetry_rollup_minutes (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    service_id      TEXT NOT NULL,
    signal_type     TEXT NOT NULL,
    bucket_start    TIMESTAMPTZ NOT NULL,
    event_count     BIGINT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, service_id, signal_type, bucket_start)
);
CREATE INDEX IF NOT EXISTS idx_rollup_tenant_bucket ON telemetry_rollup_minutes(tenant_id, bucket_start DESC);

CREATE TABLE IF NOT EXISTS log_severity_rollup_minutes (
    id           TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL,
    service_id   TEXT NOT NULL,
    severity     TEXT NOT NULL,
    bucket_start TIMESTAMPTZ NOT NULL,
    event_count  BIGINT NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, service_id, severity, bucket_start)
);

CREATE TABLE IF NOT EXISTS metric_rollup_minutes (
    id            TEXT PRIMARY KEY,
    tenant_id     TEXT NOT NULL,
    service_id    TEXT NOT NULL,
    metric_name   TEXT NOT NULL,
    bucket_start  TIMESTAMPTZ NOT NULL,
    sum_value     DOUBLE PRECISION NOT NULL DEFAULT 0,
    count_value   BIGINT NOT NULL DEFAULT 0,
    min_value     DOUBLE PRECISION,
    max_value     DOUBLE PRECISION,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, service_id, metric_name, bucket_start)
);

CREATE TABLE IF NOT EXISTS trace_latency_rollup_minutes (
    id                  TEXT PRIMARY KEY,
    tenant_id           TEXT NOT NULL,
    service_id          TEXT NOT NULL,
    bucket_start        TIMESTAMPTZ NOT NULL,
    average_duration_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    p95_duration_ms     DOUBLE PRECISION,
    span_count          BIGINT NOT NULL DEFAULT 0,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, service_id, bucket_start)
);

CREATE TABLE IF NOT EXISTS service_health_rollup_minutes (
    id           TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL,
    service_id   TEXT NOT NULL,
    bucket_start TIMESTAMPTZ NOT NULL,
    health_state TEXT NOT NULL DEFAULT 'healthy',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, service_id, bucket_start)
);

-- Alerting tables
CREATE TABLE IF NOT EXISTS alert_policies (
    id                          TEXT PRIMARY KEY,
    tenant_id                   TEXT NOT NULL,
    name                        TEXT NOT NULL,
    description                 TEXT,
    max_delivery_attempts       INT NOT NULL DEFAULT 3,
    delivery_backoff_millis     INT NOT NULL DEFAULT 200,
    escalation_interval_minutes INT NOT NULL DEFAULT 5,
    max_escalations             INT NOT NULL DEFAULT 3,
    repeat_notification_minutes INT NOT NULL DEFAULT 5,
    open_channel_types          JSONB NOT NULL DEFAULT '["webhook","slack_webhook","email"]',
    ack_channel_types           JSONB NOT NULL DEFAULT '["webhook"]',
    resolve_channel_types       JSONB NOT NULL DEFAULT '["webhook"]',
    escalation_channel_types    JSONB NOT NULL DEFAULT '["slack_webhook","webhook"]',
    active                      BOOLEAN NOT NULL DEFAULT true,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_alert_policies_tenant ON alert_policies(tenant_id);

CREATE TABLE IF NOT EXISTS alert_rules (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL,
    service_id        TEXT,
    policy_id         TEXT REFERENCES alert_policies(id),
    name              TEXT NOT NULL,
    description       TEXT,
    signal_type       TEXT NOT NULL DEFAULT 'log',
    metric_name       TEXT,
    severity          TEXT NOT NULL DEFAULT 'warning',
    aggregation       TEXT NOT NULL DEFAULT 'avg',
    comparator        TEXT NOT NULL DEFAULT '>=',
    threshold         DOUBLE PRECISION NOT NULL DEFAULT 0,
    window_minutes    INT NOT NULL DEFAULT 5,
    cooldown_minutes  INT NOT NULL DEFAULT 1,
    active            BOOLEAN NOT NULL DEFAULT true,
    last_triggered_at TIMESTAMPTZ,
    last_evaluation_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant ON alert_rules(tenant_id);

CREATE TABLE IF NOT EXISTS incidents (
    id                TEXT PRIMARY KEY,
    alert_rule_id     TEXT NOT NULL REFERENCES alert_rules(id),
    tenant_id         TEXT NOT NULL,
    service_id        TEXT,
    severity          TEXT NOT NULL DEFAULT 'warning',
    status            TEXT NOT NULL DEFAULT 'open',
    title             TEXT NOT NULL,
    summary           TEXT,
    observed_value    DOUBLE PRECISION NOT NULL DEFAULT 0,
    threshold         DOUBLE PRECISION NOT NULL DEFAULT 0,
    triggered_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    acknowledged_at   TIMESTAMPTZ,
    acknowledged_by   TEXT,
    resolved_at       TIMESTAMPTZ,
    assigned_to       TEXT,
    assigned_at       TIMESTAMPTZ,
    escalation_level  INT NOT NULL DEFAULT 0,
    escalation_count  INT NOT NULL DEFAULT 0,
    last_escalated_at TIMESTAMPTZ,
    next_escalation_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_incidents_tenant       ON incidents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status       ON incidents(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_incidents_rule_open    ON incidents(alert_rule_id, status) WHERE status = 'open';

CREATE TABLE IF NOT EXISTS incident_events (
    id          TEXT PRIMARY KEY,
    incident_id TEXT NOT NULL REFERENCES incidents(id),
    tenant_id   TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    actor_id    TEXT,
    summary     TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_incident_events_incident ON incident_events(incident_id, created_at DESC);

CREATE TABLE IF NOT EXISTS incident_comments (
    id          TEXT PRIMARY KEY,
    incident_id TEXT NOT NULL REFERENCES incidents(id),
    tenant_id   TEXT NOT NULL,
    author_id   TEXT NOT NULL,
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS notification_channels (
    id        TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name      TEXT NOT NULL,
    type      TEXT NOT NULL,
    config    JSONB NOT NULL DEFAULT '{}',
    active    BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS notification_deliveries (
    id            TEXT PRIMARY KEY,
    tenant_id     TEXT NOT NULL,
    incident_id   TEXT NOT NULL REFERENCES incidents(id),
    channel_id    TEXT NOT NULL REFERENCES notification_channels(id),
    event_type    TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    attempt_count INT NOT NULL DEFAULT 0,
    payload       JSONB,
    response      TEXT,
    delivered_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS replay_jobs (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT NOT NULL,
    service_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'pending',
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
