#!/usr/bin/env bash
set -euo pipefail

BASE_QUERY="${BASE_QUERY:-http://127.0.0.1:8084}"
BASE_TENANT="${BASE_TENANT:-http://127.0.0.1:8081}"
INTERNAL_TOKEN="${INTERNAL_TOKEN:-pulselens-internal-token}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

status_url() {
  local url="$1"
  curl -fsS "${url}" 2>/dev/null || true
}

printf "%-28s %-8s %-8s\n" "service" "health" "ready"
for row in \
  "pulselens-tenant-service|8081" \
  "pulselens-ingest-service|8082" \
  "pulselens-processing-service|8083" \
  "pulselens-query-service|8084" \
  "pulselens-alerting-service|8085" \
  "pulselens-archive-service|8086"
do
  IFS='|' read -r service port <<<"${row}"
  health="$(status_url "http://127.0.0.1:${port}/health")"
  ready="$(status_url "http://127.0.0.1:${port}/ready")"
  health_state="down"
  ready_state="down"
  [[ "${health}" == *'"status":"ok"'* ]] && health_state="ok"
  [[ "${ready}" == *'"status":"ready"'* ]] && ready_state="ready"
  printf "%-28s %-8s %-8s\n" "${service}" "${health_state}" "${ready_state}"
done

echo
echo "containers:"
docker compose ps

suffix="$(date +%s)"
email="status-admin+${suffix}@pulselens.local"
tenant_payload="$(cat <<JSON
{"name":"Status Tenant","slug":"status-${suffix}","plan":"starter","ingest_quota":1000,"retention_days":7,"admin_name":"Status Admin","admin_email":"${email}","admin_password":"password123"}
JSON
)"

tenant_resp="$(curl -fsS -X POST "${BASE_TENANT}/internal/api/v1/tenants" -H "Content-Type: application/json" -H "X-Internal-Token: ${INTERNAL_TOKEN}" -d "${tenant_payload}")"
token="$(printf '%s' "${tenant_resp}" >/dev/null; curl -fsS -X POST "${BASE_TENANT}/api/v1/auth/login" -H "Content-Type: application/json" -d "{\"email\":\"${email}\",\"password\":\"password123\"}" | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["token"])')"

echo
echo "platform dependencies:"
dependencies_json="$(curl -fsS "${BASE_QUERY}/api/v1/platform/dependencies" -H "Authorization: Bearer ${token}")"
DEPENDENCIES_JSON="${dependencies_json}" python3 - <<'PY'
import json
import os

rows = json.loads(os.environ["DEPENDENCIES_JSON"])["data"]
for row in rows:
    print(f"{row['name']}: {row['status']} ({row.get('error', '')})")
PY

echo
echo "kafka lag snapshot:"
lag_json="$(curl -fsS "${BASE_QUERY}/api/v1/platform/kafka-lag" -H "Authorization: Bearer ${token}")"
LAG_JSON="${lag_json}" python3 - <<'PY'
import json
import os

rows = json.loads(os.environ["LAG_JSON"])["data"]
print("rows=", len(rows))
for row in rows:
    print(f"{row['topic']} p{row['partition']}: lag={row['lag']}")
PY
