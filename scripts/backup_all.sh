#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_ROOT="${1:-${ROOT_DIR}/data/backups/$(date -u +%Y%m%dT%H%M%SZ)}"
mkdir -p "${BACKUP_ROOT}"

"${ROOT_DIR}/scripts/backup_postgres.sh" "${BACKUP_ROOT}" >/dev/null
"${ROOT_DIR}/scripts/backup_clickhouse.sh" "${BACKUP_ROOT}" >/dev/null
"${ROOT_DIR}/scripts/backup_minio.sh" "${BACKUP_ROOT}" >/dev/null

echo "${BACKUP_ROOT}"
