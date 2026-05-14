#!/usr/bin/env bash
set -euo pipefail

BACKUP_ROOT="${1:?backup root required}"
SQL_FILE="${BACKUP_ROOT}/postgres/pulselens.sql"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

docker compose exec -T postgres psql -U omniful -d pulselens -v ON_ERROR_STOP=1 -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
cat "${SQL_FILE}" | docker compose exec -T postgres psql -U omniful -d pulselens -v ON_ERROR_STOP=1
