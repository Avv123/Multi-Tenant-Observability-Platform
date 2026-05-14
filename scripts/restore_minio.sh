#!/usr/bin/env bash
set -euo pipefail

BACKUP_ROOT="${1:?backup root required}"
ARCHIVE_FILE="${BACKUP_ROOT}/minio/archive.tar"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

docker compose exec -T minio sh -lc 'rm -rf /data/* && mkdir -p /data'
cat "${ARCHIVE_FILE}" | docker compose exec -T minio sh -lc 'tar -C /data -xf -'
