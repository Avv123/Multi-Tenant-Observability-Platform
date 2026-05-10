#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export CONFIG_PATH

start_service() {
  local service_dir="$1"
  local extra_args="${2:-}"
  (
    cd "${ROOT_DIR}/${service_dir}"
    CONFIG_PATH=configs/config.yaml go run . ${extra_args}
  ) &
  echo $!
}

TENANT_PID="$(start_service services/pulselens-tenant-service)"
INGEST_PID="$(start_service services/pulselens-ingest-service)"
PROCESSING_PID="$(start_service services/pulselens-processing-service "-mode=all")"
QUERY_PID="$(start_service services/pulselens-query-service)"
ALERTING_PID="$(start_service services/pulselens-alerting-service "-mode=all")"
ARCHIVE_PID="$(start_service services/pulselens-archive-service "-mode=all")"

cleanup() {
  kill "${TENANT_PID}" "${INGEST_PID}" "${PROCESSING_PID}" "${QUERY_PID}" "${ALERTING_PID}" "${ARCHIVE_PID}" >/dev/null 2>&1 || true
}

trap cleanup EXIT INT TERM

echo "tenant-service pid=${TENANT_PID}"
echo "ingest-service pid=${INGEST_PID}"
echo "processing-service pid=${PROCESSING_PID}"
echo "query-service pid=${QUERY_PID}"
echo "alerting-service pid=${ALERTING_PID}"
echo "archive-service pid=${ARCHIVE_PID}"
echo "Press Ctrl+C to stop all services."

wait
