# Local Development

## Supported Runtime

PulseLens is run locally through `docker compose` only. The public startup flow is:

1. `bash scripts/check_required_ports.sh`
2. `docker compose up -d --build`
3. `bash scripts/wait_for_compose_stack.sh`
4. open `http://127.0.0.1:3000`

Use `bash scripts/local_status.sh` to inspect service health, readiness, dependency state, and Kafka lag.

## Ports

### Product services

- UI: `3000`
- tenant-service: `8081`
- ingest-service: `8082`
- processing-service: `8083`
- query-service: `8084`
- alerting-service: `8085`
- archive-service: `8086`

### Dependencies

- Postgres: `5433`
- Redis: `6381`
- Kafka: `9093`
- ClickHouse: `8123`
- MinIO API: `9010`
- MinIO Console: `9011`
- MailHog SMTP: `1026`
- MailHog UI/API: `8026`

## Runtime Notes

- backend services expose both `/health` and `/ready`
- `/health` means the service process is alive
- `/ready` means the service can serve traffic with its required dependencies
- port conflicts are reported explicitly and never auto-killed
- host-run binaries are not part of the supported flow

## Verification

### Basic checks

- `python3 scripts/test_end_to_end.py`
- `bash scripts/curl_smoke.sh`
- `bash scripts/local_status.sh`

### Performance and failure

- `python3 scripts/test_load.py`
- `python3 scripts/test_sustained_load.py`
- `python3 scripts/test_notification_failures.py`

### Recovery and resilience

- `bash scripts/backup_all.sh`
- `bash scripts/restore_all.sh <backup-root>`
- `python3 scripts/test_backup_restore.py`
- `python3 scripts/test_redis_outage.py`
- `python3 scripts/test_clickhouse_outage.py`
- `python3 scripts/test_kafka_outage.py`
- `python3 scripts/test_postgres_outage.py`
- `python3 scripts/test_minio_outage.py`
- `python3 scripts/test_multi_dependency_outage.py`

### One-shot hardening

- `bash scripts/run_full_validation.sh`

## Related Runbooks

- implementation inventory: `docs/08-implemented-system/README.md`
- backup and restore: `docs/11-backup-restore/README.md`
- chaos drills: `docs/12-chaos-drills/README.md`
- security operations: `docs/13-security-operations/README.md`
