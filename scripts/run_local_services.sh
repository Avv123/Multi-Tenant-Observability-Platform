#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT_DIR}/bin"

export CONFIG_PATH
export CONFIG_INLINE

start_service() {
  local service_dir="$1"
  local binary_name="$2"
  local extra_args="${3:-}"
  (
    cd "${ROOT_DIR}/${service_dir}"
    CONFIG_INLINE="$(cat configs/config.yaml)" "${BIN_DIR}/${binary_name}" ${extra_args}
  ) &
  echo $!
}

"${ROOT_DIR}/scripts/build_local_binaries.sh"

TENANT_PID="$(start_service services/pulselens-tenant-service pulselens-tenant-service)"
INGEST_PID="$(start_service services/pulselens-ingest-service pulselens-ingest-service)"
PROCESSING_PID="$(start_service services/pulselens-processing-service pulselens-processing-service "-mode=all")"
QUERY_PID="$(start_service services/pulselens-query-service pulselens-query-service)"
ALERTING_PID="$(start_service services/pulselens-alerting-service pulselens-alerting-service "-mode=all")"
ARCHIVE_PID="$(start_service services/pulselens-archive-service pulselens-archive-service "-mode=all")"
(
  cd "${ROOT_DIR}/services/pulselens-ui"
  npm run preview -- --host 127.0.0.1 --port 3000
) &
UI_PID="$!"

cleanup() {
  kill "${TENANT_PID}" "${INGEST_PID}" "${PROCESSING_PID}" "${QUERY_PID}" "${ALERTING_PID}" "${ARCHIVE_PID}" "${UI_PID}" >/dev/null 2>&1 || true
}

trap cleanup EXIT INT TERM

echo "tenant-service pid=${TENANT_PID}"
echo "ingest-service pid=${INGEST_PID}"
echo "processing-service pid=${PROCESSING_PID}"
echo "query-service pid=${QUERY_PID}"
echo "alerting-service pid=${ALERTING_PID}"
echo "archive-service pid=${ARCHIVE_PID}"
echo "ui-preview pid=${UI_PID}"
echo "Press Ctrl+C to stop all services."

wait
