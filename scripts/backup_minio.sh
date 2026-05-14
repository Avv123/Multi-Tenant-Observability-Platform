#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_ROOT="${1:-${ROOT_DIR}/data/backups/$(date -u +%Y%m%dT%H%M%SZ)}"
TARGET_DIR="${BACKUP_ROOT}/minio"
mkdir -p "${TARGET_DIR}"
cd "${ROOT_DIR}"

docker compose exec -T minio sh -lc 'tar -C /data -cf - .' > "${TARGET_DIR}/archive.tar"

cat > "${TARGET_DIR}/metadata.json" <<JSON
{
  "service": "minio",
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
JSON

echo "${TARGET_DIR}"
