#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_DIR="${ROOT_DIR}/data/benchmark-reports/$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "${REPORT_DIR}"

run() {
  local name="$1"
  shift
  echo "== ${name} =="
  "$@"
}

run "check_required_ports" bash "${ROOT_DIR}/scripts/check_required_ports.sh"
run "compose_up" docker compose up -d --build
run "wait_for_compose_stack" bash "${ROOT_DIR}/scripts/wait_for_compose_stack.sh"
run "end_to_end" python3 "${ROOT_DIR}/scripts/test_end_to_end.py" > "${REPORT_DIR}/end_to_end.json"
run "load" python3 "${ROOT_DIR}/scripts/test_load.py" --json-out "${REPORT_DIR}/load.json"
run "sustained_load" python3 "${ROOT_DIR}/scripts/test_sustained_load.py" --json-out "${REPORT_DIR}/soak.json"
run "notification_failures" python3 "${ROOT_DIR}/scripts/test_notification_failures.py" > "${REPORT_DIR}/notification_failures.json"
run "redis_outage" python3 "${ROOT_DIR}/scripts/test_redis_outage.py" > "${REPORT_DIR}/redis_outage.json"
run "clickhouse_outage" python3 "${ROOT_DIR}/scripts/test_clickhouse_outage.py" > "${REPORT_DIR}/clickhouse_outage.json"
run "kafka_outage" python3 "${ROOT_DIR}/scripts/test_kafka_outage.py" > "${REPORT_DIR}/kafka_outage.json"
run "postgres_outage" python3 "${ROOT_DIR}/scripts/test_postgres_outage.py" > "${REPORT_DIR}/postgres_outage.json"
run "benchmark_report" python3 "${ROOT_DIR}/scripts/generate_benchmark_report.py" "${REPORT_DIR}" "${REPORT_DIR}/load.json" "${REPORT_DIR}/soak.json"

echo
echo "almost production local check complete"
echo "report directory: ${REPORT_DIR}"
