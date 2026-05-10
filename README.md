# PulseLens

`PulseLens` is a local-first, multi-tenant observability platform built in the same broad Go service style used across your Omniful services.

The implemented stack now includes:

- `pulselens-tenant-service`
- `pulselens-ingest-service`
- `pulselens-processing-service`
- `pulselens-query-service`
- `pulselens-alerting-service`
- `pulselens-archive-service`
- `pulselens-ui` as a React application
- `libs/pulselens-platform` as the local shared package layer

## Quick Start

1. `bash scripts/setup_local_infra.sh`
2. `cd services/pulselens-ui && npm install && npm run build`
3. `bash scripts/run_local_services.sh`
4. Open `http://127.0.0.1:8084`

## Reproducible Runtime

The repo now also includes:

- `docker-compose.yml` for the full local stack
- per-service `Dockerfile` definitions
- optional Helm assets under `deploy/helm/pulselens`

## Verification Paths

- `python3 scripts/smoke_test.py`
- `bash scripts/curl_smoke.sh`
- `python3 scripts/load_test.py`
- `python3 scripts/soak_test.py`
- `python3 scripts/failure_drill.py`

Both flows were run successfully against the current implementation.

## V2 Additions

The platform now also includes:

- ClickHouse hot-store support for logs, metrics, and traces
- ClickHouse materialized minute rollups and ClickHouse-first aggregate queries
- ClickHouse-first query path with PostgreSQL fallback
- Redis query caching with tenant-scoped invalidation
- minute rollups for overview, service-health, log severity, metric series, and trace latency
- delayed retry orchestration
- tenant-aware retention cleanup
- object-store-backed archive and replay jobs
- Snowflake-style sortable IDs
- service heartbeats and runtime visibility
- Redis-backed backpressure controls
- retention cleanup runs
- saved queries
- saved dashboards
- incident acknowledge, resolve, assign, and escalation workflow
- incident comments
- webhook, Slack-webhook-style, and email notification delivery
- tenant user management with RBAC enforcement
- local concurrent load-test harness
- soak and failure-drill scripts
- Playwright browser E2E validation

## Document Map

- [Overview](docs/01-overview/README.md)
- [System Architecture](docs/02-system-architecture/README.md)
- [Implementation Style](docs/03-implementation-style/README.md)
- [Data And Flows](docs/04-data-and-flows/README.md)
- [Build Plan](docs/05-build-plan/README.md)
- [Local Development](docs/06-local-development/README.md)
- [AWS Deployment](docs/07-aws-deployment/README.md)
- [Implemented System](docs/08-implemented-system/README.md)
- [System Design Notes](docs/09-system-design/README.md)
