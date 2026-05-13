#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT_DIR}/bin"

mkdir -p "${BIN_DIR}"

build_service() {
  local service_dir="$1"
  local output_name="$2"
  (
    cd "${ROOT_DIR}/${service_dir}"
    go build -o "${BIN_DIR}/${output_name}" .
  )
}

build_service services/pulselens-tenant-service pulselens-tenant-service
build_service services/pulselens-ingest-service pulselens-ingest-service
build_service services/pulselens-processing-service pulselens-processing-service
build_service services/pulselens-query-service pulselens-query-service
build_service services/pulselens-alerting-service pulselens-alerting-service
build_service services/pulselens-archive-service pulselens-archive-service

if [[ -f "${ROOT_DIR}/services/pulselens-ui/package.json" ]]; then
  (
    cd "${ROOT_DIR}/services/pulselens-ui"
    if [[ ! -d node_modules ]]; then
      npm install
    fi
    npm run build
  )
fi

echo "built binaries into ${BIN_DIR}"
