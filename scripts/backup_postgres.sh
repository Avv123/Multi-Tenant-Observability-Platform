#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_ROOT="${1:-${ROOT_DIR}/data/backups/$(date -u +%Y%m%dT%H%M%SZ)}"
TARGET_DIR="${BACKUP_ROOT}/postgres"
mkdir -p "${TARGET_DIR}"
cd "${ROOT_DIR}"

docker compose exec -T postgres pg_dump -U pulselens -d pulselens --no-owner --no-privileges >"${TARGET_DIR}/pulselens.sql"
cat > "${TARGET_DIR}/metadata.json" <<JSON
{
  "database": "pulselens",
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "service": "postgres"
}
JSON

echo "${TARGET_DIR}"
