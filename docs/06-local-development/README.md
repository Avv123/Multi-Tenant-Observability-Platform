# Local Development

## Why Local-First

The full platform should run locally because:

- the goal is zero recurring cost
- all critical backend paths can still be proven
- Docker Compose is enough for this scale
- you can demonstrate the real system without cloud spend

## Local Stack

Run locally with:

- `tenant-service`
- `ingest-service`
- `processing-service`
- `query-service`
- `ui`
- `Kafka`
- `Zookeeper` or KRaft mode depending on chosen image
- `Redis`
- `ClickHouse`
- `PostgreSQL`
- `Prometheus`
- `Grafana`

## Local Topology

```mermaid
flowchart LR
UI[React UI] --> QRY[Query Service]
UI --> TEN[Tenant Service]
TEN --> PG[PostgreSQL]
ING[Ingest Service] --> K[Kafka]
ING --> R[Redis]
ING --> TEN
K --> PROC[Processing Service]
PROC --> CH[ClickHouse]
PROC --> R
PROC --> S3M[Local object storage or filesystem archive]
QRY --> CH
QRY --> R
PROM[Prometheus] --> ING
PROM --> PROC
PROM --> QRY
GRAF[Grafana] --> PROM
```

## Recommended Local Storage Adjustments

For zero-cost local development:

- use local filesystem or MinIO-compatible storage instead of real S3
- keep archive layout S3-like so future migration is easy

Why:

- preserves production semantics without cloud cost

## Local Demo Dataset

Prepare at least:

- `tenant-alpha`
- `tenant-beta`
- `checkout-service`
- `catalog-service`
- `shipping-service`

And generate:

- normal request logs
- warning/error bursts
- a few metrics with labels
- traces spanning multiple services

Why:

- demo data should prove isolation and distributed flow

## Local Acceptance Demo

The reviewer should be able to do this:

1. Create or seed tenants
2. Create API keys
3. Send telemetry for tenant A and tenant B
4. Verify both are ingested
5. Trigger retries and one DLQ case
6. Search logs by tenant and service
7. Open a trace by trace ID
8. View metrics rollups in the UI
9. Inspect Prometheus/Grafana for internal platform health

## Local Test Layers

### Unit Tests

For:

- request validation
- service logic
- dedupe helpers
- rate limit logic
- error mapping

### Integration Tests

For:

- Kafka publish and consume
- Redis dedupe and rate limit
- ClickHouse inserts and queries
- PostgreSQL control-plane flows

### End-To-End Tests

For:

- full ingest to query journey
- cross-tenant isolation
- retry and DLQ scenarios

## Load Testing Plan

Use `k6`.

Scenarios:

- constant low traffic
- burst traffic
- multi-tenant mixed traffic
- duplicate replay traffic
- degraded worker/storage conditions

Metrics to capture:

- ingest p95
- query p95
- Kafka lag
- worker throughput
- DLQ rate
- retry success rate

## Why This Matters In Review

Interviewers trust real measured behavior more than claims. A local benchmark with a believable setup is far better than a cloud URL with shallow logic.

