# Backup And Restore

## Commands
- `bash scripts/backup_all.sh`
- `bash scripts/restore_all.sh <backup-root>`
- `python3 scripts/test_backup_restore.py`

## Coverage
- Postgres control-plane data
- ClickHouse hot telemetry and rollups
- MinIO archive objects

## Restore Drill
The restore drill:
1. bootstraps tenant and telemetry data
2. creates backups
3. wipes project runtime state
4. restores all backup targets
5. validates login, telemetry, dashboards, and archive stats
