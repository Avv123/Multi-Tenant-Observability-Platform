#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT_DIR}/bin"
LOG_DIR="${ROOT_DIR}/.tmp/local-services"

mkdir -p "${LOG_DIR}"

export CONFIG_PATH
export CONFIG_INLINE

declare -a PIDS=()

wait_for_port() {
  local port="$1"
  local retries="${2:-20}"
  for _ in $(seq 1 "${retries}"); do
    if lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  return 1
}

cleanup() {
  for pid in "${PIDS[@]:-}"; do
    kill "${pid}" >/dev/null 2>&1 || true
  done
}

trap cleanup EXIT INT TERM

start_service() {
  local service_name="$1"
  local service_dir="$2"
  local binary_name="$3"
  local port="$4"
  local extra_args="${5:-}"
  local log_path="${LOG_DIR}/${service_name}.log"

  : > "${log_path}"
  (
    cd "${ROOT_DIR}/${service_dir}"
    CONFIG_INLINE="$(cat configs/config.yaml)" "${BIN_DIR}/${binary_name}" ${extra_args}
  ) >"${log_path}" 2>&1 &
  local pid=$!
  PIDS+=("${pid}")

  sleep 1
  if ! kill -0 "${pid}" >/dev/null 2>&1; then
    echo "failed: ${service_name} exited immediately"
    tail -n 40 "${log_path}" || true
    return 1
  fi

  if ! wait_for_port "${port}"; then
    echo "failed: ${service_name} did not bind port ${port}"
    tail -n 40 "${log_path}" || true
    return 1
  fi

  echo "started: ${service_name} pid=${pid} port=${port} log=${log_path}"
}

start_ui_preview() {
  local log_path="${LOG_DIR}/pulselens-ui-preview.log"
  : > "${log_path}"
  (
    cd "${ROOT_DIR}/services/pulselens-ui"
    npm run build >/dev/null
    npm run preview -- --host 127.0.0.1 --port 3000
  ) >"${log_path}" 2>&1 &
  local pid=$!
  PIDS+=("${pid}")

  sleep 1
  if ! kill -0 "${pid}" >/dev/null 2>&1; then
    echo "failed: pulselens-ui-preview exited immediately"
    tail -n 40 "${log_path}" || true
    return 1
  fi

  if ! wait_for_port 3000; then
    echo "failed: pulselens-ui-preview did not bind port 3000"
    tail -n 40 "${log_path}" || true
    return 1
  fi

  echo "started: pulselens-ui-preview pid=${pid} port=3000 log=${log_path}"
}

"${ROOT_DIR}/scripts/build_local_binaries.sh"

start_service "pulselens-tenant-service" "services/pulselens-tenant-service" "pulselens-tenant-service" 8081
start_service "pulselens-ingest-service" "services/pulselens-ingest-service" "pulselens-ingest-service" 8082
start_service "pulselens-processing-service" "services/pulselens-processing-service" "pulselens-processing-service" 8083 "-mode=all"
start_service "pulselens-query-service" "services/pulselens-query-service" "pulselens-query-service" 8084
start_service "pulselens-alerting-service" "services/pulselens-alerting-service" "pulselens-alerting-service" 8085 "-mode=all"
start_service "pulselens-archive-service" "services/pulselens-archive-service" "pulselens-archive-service" 8086 "-mode=all"
start_ui_preview

echo "all services started"
echo "logs directory: ${LOG_DIR}"
echo "press Ctrl+C to stop all services"

wait
