#!/usr/bin/env bash
set -euo pipefail

BACKUP_ROOT="${1:?backup root required}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

"${ROOT_DIR}/scripts/restore_postgres.sh" "${BACKUP_ROOT}"
"${ROOT_DIR}/scripts/restore_clickhouse.sh" "${BACKUP_ROOT}"
"${ROOT_DIR}/scripts/restore_minio.sh" "${BACKUP_ROOT}"
