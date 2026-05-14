#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

services=(postgres redis kafka clickhouse minio mailhog tenant-service ingest-service processing-service query-service alerting-service archive-service ui)

wait_for_container_health() {
  local service="$1"
  local retries="${2:-90}"
  for _ in $(seq 1 "${retries}"); do
    status="$(docker compose ps --format json 2>/dev/null | python3 -c 'import json,sys
rows=[json.loads(line) for line in sys.stdin if line.strip()]
target=sys.argv[1]
for row in rows:
    if row.get("Service")==target:
        health=row.get("Health") or row.get("State") or ""
        print(health)
        break
' "${service}" || true)"
    if [[ "${status}" == "healthy" || "${status}" == "running" ]]; then
      return 0
    fi
    sleep 2
  done
  echo "container health timeout: ${service}" >&2
  docker compose logs --tail=60 "${service}" >&2 || true
  return 1
}

wait_for_http_ready() {
  local name="$1"
  local port="$2"
  local retries="${3:-90}"
  for _ in $(seq 1 "${retries}"); do
    if curl -fsS "http://127.0.0.1:${port}/ready" | grep -q '"status":"ready"'; then
      return 0
    fi
    sleep 2
  done
  echo "readiness timeout: ${name}" >&2
  docker compose logs --tail=80 "${name}" >&2 || true
  return 1
}

for service in "${services[@]}"; do
  wait_for_container_health "${service}"
done

wait_for_http_ready tenant-service 8081
wait_for_http_ready ingest-service 8082
wait_for_http_ready processing-service 8083
wait_for_http_ready query-service 8084
wait_for_http_ready alerting-service 8085
wait_for_http_ready archive-service 8086

if ! curl -fsS "http://127.0.0.1:3000/" >/dev/null; then
  echo "ui readiness timeout" >&2
  docker compose logs --tail=80 ui >&2 || true
  exit 1
fi

echo "compose stack ready"
