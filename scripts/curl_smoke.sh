#!/usr/bin/env bash
set -euo pipefail

BASE_TENANT="http://127.0.0.1:8081"
BASE_INGEST="http://127.0.0.1:8082"
BASE_QUERY="http://127.0.0.1:8084"
BASE_ALERTING="http://127.0.0.1:8085"
BASE_ARCHIVE="http://127.0.0.1:8086"
BASE_UI="http://127.0.0.1:3000"
INTERNAL_TOKEN="pulselens-internal-token"
SUFFIX="$(date +%s)"
EMAIL="curl-admin+${SUFFIX}@pulselens.local"
SLUG="curl-${SUFFIX}"

extract_json_field() {
  python3 -c 'import json,sys; data=json.load(sys.stdin); print(eval(sys.argv[1], {"__builtins__": {}}, {"data": data, "len": len}))' "$1"
}

create_tenant_response="$(curl -sS -X POST "${BASE_TENANT}/internal/api/v1/tenants" \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: ${INTERNAL_TOKEN}" \
  -d "{
    \"name\": \"Curl Tenant\",
    \"slug\": \"${SLUG}\",
    \"plan\": \"starter\",
    \"ingest_quota\": 1000,
    \"retention_days\": 7,
    \"admin_name\": \"Curl Admin\",
    \"admin_email\": \"${EMAIL}\",
    \"admin_password\": \"password123\"
  }")"

tenant_id="$(printf '%s' "${create_tenant_response}" | extract_json_field "data['data']['tenant']['id']")"

login_response="$(curl -sS -X POST "${BASE_TENANT}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${EMAIL}\",
    \"password\": \"password123\"
  }")"

token="$(printf '%s' "${login_response}" | extract_json_field "data['data']['token']")"

create_service_response="$(curl -sS -X POST "${BASE_TENANT}/admin/api/v1/tenants/${tenant_id}/services" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d '{
    "name": "curl-api",
    "environment": "local",
    "tags": {"team":"platform","source":"curl-smoke"}
  }')"

service_id="$(printf '%s' "${create_service_response}" | extract_json_field "data['data']['id']")"

create_key_response="$(curl -sS -X POST "${BASE_TENANT}/admin/api/v1/api-keys" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d "{
    \"tenant_id\": \"${tenant_id}\",
    \"service_id\": \"${service_id}\",
    \"name\": \"curl-key\",
    \"scopes\": [\"ingest\", \"query\", \"admin\"]
  }")"

api_key="$(printf '%s' "${create_key_response}" | extract_json_field "data['data']['key']")"

viewer_email="viewer-${SUFFIX}@pulselens.local"
curl -sS -X POST "${BASE_TENANT}/admin/api/v1/tenants/${tenant_id}/users" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d "{
    \"name\": \"Viewer User\",
    \"email\": \"${viewer_email}\",
    \"password\": \"viewer-pass\",
    \"role\": \"viewer\"
  }" >/dev/null

curl -sS -X POST "${BASE_INGEST}/api/v1/ingest" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${api_key}" \
  -d '{
    "events": [
      {"event_type":"log","severity":"error","payload":{"message":"curl checkout failure","logger":"curl-smoke"}},
      {"event_type":"metric","payload":{"metric_name":"curl_latency_ms","value":180.5,"unit":"ms"}},
      {"event_type":"trace","trace_id":"curl-trace-1","payload":{"span_id":"root","parent_span_id":"","operation":"checkout","status":"ok","start_time":"2025-01-01T00:00:00Z","end_time":"2025-01-01T00:00:00.180Z"}}
    ]
  }' >/dev/null

curl -sS -X POST "${BASE_ALERTING}/api/v1/alert-rules" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d "{
    \"service_id\": \"${service_id}\",
    \"name\": \"Curl Error Burst\",
    \"signal_type\": \"log\",
    \"severity\": \"error\",
    \"aggregation\": \"count\",
    \"comparator\": \">=\",
    \"threshold\": 1,
    \"window_minutes\": 10,
    \"cooldown_minutes\": 1
  }" >/dev/null

curl -sS -X POST "${BASE_ALERTING}/api/v1/notification-channels" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d '{
    "name":"Curl Log Channel",
    "type":"log",
    "config":{"destination":"local-log"}
  }' >/dev/null

curl -sS -X POST "${BASE_ALERTING}/api/v1/notification-channels" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d "{
    \"name\":\"Curl Email Channel\",
    \"type\":\"email\",
    \"config\":{\"to\":[\"alerts-${SUFFIX}@pulselens.local\"],\"subject_prefix\":\"[Curl]\"}
  }" >/dev/null

curl -sS -X POST "${BASE_QUERY}/api/v1/saved-queries" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d "{
    \"name\":\"Curl Error Query\",
    \"query_type\":\"logs\",
    \"definition\":{\"service_id\":\"${service_id}\",\"severity\":\"error\",\"limit\":20}
  }" >/dev/null

curl -sS -X POST "${BASE_QUERY}/api/v1/dashboards" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d '{
    "name":"Curl Dashboard",
    "description":"curl smoke dashboard",
    "layout":{"columns":2},
    "widgets":[{"type":"stat","metric":"log_count"}]
  }' >/dev/null

sleep 4

curl -sS -X POST "${BASE_ARCHIVE}/api/v1/replay-jobs" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${token}" \
  -d "{
    \"service_id\":\"${service_id}\",
    \"event_type\":\"log\",
    \"start_time\":\"2020-01-01T00:00:00Z\",
    \"end_time\":\"2100-01-01T00:00:00Z\"
  }" >/dev/null

sleep 5

overview="$(curl -sS "${BASE_QUERY}/api/v1/overview" -H "Authorization: Bearer ${token}")"
platform_overview="$(curl -sS "${BASE_QUERY}/api/v1/platform/overview?limit=10" -H "Authorization: Bearer ${token}")"
platform_dependencies="$(curl -sS "${BASE_QUERY}/api/v1/platform/dependencies" -H "Authorization: Bearer ${token}")"
kafka_lag="$(curl -sS "${BASE_QUERY}/api/v1/platform/kafka-lag" -H "Authorization: Bearer ${token}")"
log_severity="$(curl -sS "${BASE_QUERY}/api/v1/analytics/log-severity?limit=20" -H "Authorization: Bearer ${token}")"
metric_series="$(curl -sS "${BASE_QUERY}/api/v1/analytics/metric-series?limit=20&metric_name=curl_latency_ms" -H "Authorization: Bearer ${token}")"
trace_latency="$(curl -sS "${BASE_QUERY}/api/v1/analytics/trace-latency?limit=20" -H "Authorization: Bearer ${token}")"
incidents="$(curl -sS "${BASE_ALERTING}/api/v1/incidents" -H "Authorization: Bearer ${token}")"
saved_queries="$(curl -sS "${BASE_QUERY}/api/v1/saved-queries" -H "Authorization: Bearer ${token}")"
dashboards="$(curl -sS "${BASE_QUERY}/api/v1/dashboards" -H "Authorization: Bearer ${token}")"
replay_jobs="$(curl -sS "${BASE_ARCHIVE}/api/v1/replay-jobs" -H "Authorization: Bearer ${token}")"
archive_stats="$(curl -sS "${BASE_ARCHIVE}/api/v1/archive/stats" -H "Authorization: Bearer ${token}")"
users="$(curl -sS "${BASE_TENANT}/admin/api/v1/tenants/${tenant_id}/users" -H "Authorization: Bearer ${token}")"
ui_root="$(curl -sS "${BASE_UI}/")"

printf 'tenant_id=%s\n' "${tenant_id}"
printf 'service_id=%s\n' "${service_id}"
printf 'log_count=%s\n' "$(printf '%s' "${overview}" | extract_json_field "data['data']['log_count']")"
printf 'runtime_count=%s\n' "$(printf '%s' "${platform_overview}" | extract_json_field "len(data['data']['runtime'])")"
printf 'cleanup_run_count=%s\n' "$(printf '%s' "${platform_overview}" | extract_json_field "len(data['data']['cleanup_runs'])")"
printf 'dependency_count=%s\n' "$(printf '%s' "${platform_dependencies}" | extract_json_field "len(data['data'])")"
printf 'kafka_lag_rows=%s\n' "$(printf '%s' "${kafka_lag}" | extract_json_field "len(data['data'])")"
printf 'log_severity_count=%s\n' "$(printf '%s' "${log_severity}" | extract_json_field "len(data['data'])")"
printf 'metric_series_count=%s\n' "$(printf '%s' "${metric_series}" | extract_json_field "len(data['data'])")"
printf 'trace_latency_count=%s\n' "$(printf '%s' "${trace_latency}" | extract_json_field "len(data['data'])")"
printf 'incident_count=%s\n' "$(printf '%s' "${incidents}" | extract_json_field "len(data['data'])")"
printf 'saved_query_count=%s\n' "$(printf '%s' "${saved_queries}" | extract_json_field "len(data['data'])")"
printf 'dashboard_count=%s\n' "$(printf '%s' "${dashboards}" | extract_json_field "len(data['data'])")"
printf 'replay_job_count=%s\n' "$(printf '%s' "${replay_jobs}" | extract_json_field "len(data['data'])")"
printf 'archive_object_count=%s\n' "$(printf '%s' "${archive_stats}" | extract_json_field "data['data']['archive_object_count']")"
printf 'user_count=%s\n' "$(printf '%s' "${users}" | extract_json_field "len(data['data'])")"
printf 'ui_served=%s\n' "$(printf '%s' "${ui_root}" | python3 -c 'import sys; print("PulseLens" in sys.stdin.read())')"
