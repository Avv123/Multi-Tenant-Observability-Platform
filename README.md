# PulseLens

PulseLens is a local-first, multi-tenant observability platform that ingests, processes, stores, and queries logs, metrics, and traces on a single machine. It includes dashboards, incidents, alert delivery, archive/replay, backup/restore, chaos drills, and operational readiness checks, all packaged around a Compose-owned runtime.

## What It Does

- accepts telemetry for multiple tenants through one platform
- separates control-plane concerns from telemetry processing and query flows
- supports log search, metric series, trace lookup, and service health views
- includes saved queries, saved dashboards, and a dashboard builder
- supports alert rules, incidents, assignments, comments, and notification delivery
- archives raw telemetry to object storage and can replay archived data
- ships with smoke, load, soak, failure, backup/restore, and chaos validation flows

## Architecture

- `pulselens-tenant-service`
  - authentication, tenants, users, roles, services, API keys, audit logs
- `pulselens-ingest-service`
  - API-key-authenticated telemetry ingestion, validation, rate limits, quotas, Kafka publish
- `pulselens-processing-service`
  - Kafka consumers, dedupe, retries, retention, ClickHouse writes, archive writes
- `pulselens-query-service`
  - logs, metrics, traces, rollups, dashboards, saved queries, platform health
- `pulselens-alerting-service`
  - alert rules, incidents, escalation, comments, delivery tracking
- `pulselens-archive-service`
  - archive stats, replay jobs, replay worker flows
- `pulselens-ui`
  - React UI for investigation, dashboards, incidents, alerts, users, and runtime visibility

## Stack

- Go backend services
- React + Vite frontend
- PostgreSQL for control-plane state
- Redis for cache, locks, heartbeats, and backpressure state
- Redpanda/Kafka for asynchronous telemetry processing
- ClickHouse for hot telemetry reads and rollups
- MinIO for archive and replay storage
- MailHog for local SMTP testing
- Docker Compose for the supported local runtime
- optional Helm + `k3d` validation for packaging checks

## Quick Start

1. `bash scripts/check_required_ports.sh`
2. `docker compose up -d --build`
3. `bash scripts/wait_for_compose_stack.sh`
4. open `http://127.0.0.1:3000`

## Ports

- UI: `3000`
- tenant-service: `8081`
- ingest-service: `8082`
- processing-service: `8083`
- query-service: `8084`
- alerting-service: `8085`
- archive-service: `8086`
- Postgres: `5433`
- Redis: `6381`
- Kafka: `9093`
- ClickHouse: `8123`
- MinIO API: `9010`
- MinIO Console: `9011`
- MailHog SMTP: `1026`
- MailHog UI/API: `8026`

## Validation

The repo includes:

- smoke verification
- load and soak tests with threshold gates
- notification failure drills
- backup and restore drill
- Redis, ClickHouse, Kafka, Postgres, and MinIO chaos drills
- optional Helm render and local Kubernetes dry-run/smoke validation

Use `bash scripts/run_full_validation.sh` for the full local hardening flow. Public runbooks are linked below.

## Runtime Notes

- the only supported local runtime is `docker compose`
- host-run binaries are not part of the supported project flow
- PulseLens owns its declared local ports
- port conflicts are reported explicitly and never auto-killed

## Docs

- [Overview](docs/01-overview/README.md)
- [Local Development](docs/06-local-development/README.md)
- [Implemented System](docs/08-implemented-system/README.md)
- [System Design Notes](docs/09-system-design/README.md)
- [Local SLOs](docs/10-local-slos/README.md)
- [Backup And Restore](docs/11-backup-restore/README.md)
- [Chaos Drills](docs/12-chaos-drills/README.md)
- [Security Operations](docs/13-security-operations/README.md)
- [Kubernetes Validation Notes](deploy/k8s/README.md)
