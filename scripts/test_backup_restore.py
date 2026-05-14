#!/usr/bin/env python3
import json
import subprocess
import time
import urllib.request

BASE_TENANT = "http://127.0.0.1:8081"
BASE_INGEST = "http://127.0.0.1:8082"
BASE_QUERY = "http://127.0.0.1:8084"
BASE_ARCHIVE = "http://127.0.0.1:8086"
INTERNAL_TOKEN = "pulselens-internal-token"


def req(url, method="GET", body=None, headers=None):
    headers = headers or {}
    payload = None
    if body is not None:
        payload = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, method=method, data=payload, headers=headers)
    with urllib.request.urlopen(request, timeout=20) as response:
        return json.loads(response.read().decode())


def wait_until(predicate, timeout_seconds=30, interval_seconds=1):
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        value = predicate()
        if value:
            return value
        time.sleep(interval_seconds)
    return None


def bootstrap():
    suffix = str(int(time.time()))
    email = f"restore-admin+{suffix}@pulselens.local"
    tenant = req(
        BASE_TENANT + "/internal/api/v1/tenants",
        "POST",
        {
            "name": "Restore Drill Tenant",
            "slug": f"restore-{suffix}",
            "plan": "starter",
            "ingest_quota": 5000,
            "retention_days": 7,
            "admin_name": "Restore Admin",
            "admin_email": email,
            "admin_password": "password123",
        },
        {"X-Internal-Token": INTERNAL_TOKEN},
    )["data"]["tenant"]
    login = req(BASE_TENANT + "/api/v1/auth/login", "POST", {"email": email, "password": "password123"})["data"]
    service = req(
        BASE_TENANT + f"/admin/api/v1/tenants/{tenant['id']}/services",
        "POST",
        {"name": "restore-api", "environment": "local", "tags": {"suite": "restore"}},
        {"Authorization": f"Bearer {login['token']}"},
    )["data"]
    api_key = req(
        BASE_TENANT + "/admin/api/v1/api-keys",
        "POST",
        {"tenant_id": tenant["id"], "service_id": service["id"], "name": "restore-key", "scopes": ["ingest", "query", "admin"]},
        {"Authorization": f"Bearer {login['token']}"},
    )["data"]["key"]
    req(
        BASE_QUERY + "/api/v1/dashboards",
        "POST",
        {"name": "Restore Dashboard", "description": "restore drill", "layout": {"columns": 2}, "widgets": [{"type": "stat", "metric": "log_count"}]},
        {"Authorization": f"Bearer {login['token']}"},
    )
    for event in [
        {"event_type": "log", "severity": "error", "payload": {"message": "restore-log", "logger": "restore"}},
        {"event_type": "metric", "payload": {"metric_name": "restore_metric_ms", "value": 101}},
        {
            "event_type": "trace",
            "trace_id": f"restore-trace-{suffix}",
            "payload": {"span_id": "root", "parent_span_id": "", "operation": "restore", "status": "ok", "start_time": "2025-01-01T00:00:00Z", "end_time": "2025-01-01T00:00:00.101Z"},
        },
    ]:
        req(BASE_INGEST + "/api/v1/ingest", "POST", {"events": [event]}, {"X-API-Key": api_key})
    wait_until(lambda: req(BASE_QUERY + "/api/v1/overview", headers={"Authorization": f"Bearer {login['token']}"})["data"]["log_count"] >= 1)
    replay_job = req(
        BASE_ARCHIVE + "/api/v1/replay-jobs",
        "POST",
        {
            "service_id": service["id"],
            "event_type": "log",
            "start_time": "2024-12-31T23:59:00Z",
            "end_time": "2025-01-01T00:10:00Z",
        },
        {"Authorization": f"Bearer {login['token']}"},
    )["data"]
    return tenant, service, api_key, login["token"], replay_job["id"], email


def main():
    tenant, service, api_key, token, _, email = bootstrap()
    backup_root = subprocess.check_output(["bash", "scripts/backup_all.sh"], text=True).strip()
    subprocess.check_call(["bash", "scripts/wipe_runtime_data.sh"])
    subprocess.check_call(["bash", "scripts/restore_all.sh", backup_root])
    time.sleep(5)

    relogin = req(BASE_TENANT + "/api/v1/auth/login", "POST", {"email": email, "password": "password123"})
    restored_token = relogin["data"]["token"]
    overview = wait_until(lambda: req(BASE_QUERY + "/api/v1/overview", headers={"Authorization": f"Bearer {restored_token}"})["data"], timeout_seconds=30)
    dashboards = req(BASE_QUERY + "/api/v1/dashboards", headers={"Authorization": f"Bearer {restored_token}"})["data"]
    replay_stats = req(BASE_ARCHIVE + "/api/v1/archive/stats", headers={"Authorization": f"Bearer {restored_token}"})["data"]
    req(BASE_INGEST + "/api/v1/ingest", "POST", {"events": [{"event_type": "log", "severity": "info", "payload": {"message": "post-restore-ingest", "logger": "restore"}}]}, {"X-API-Key": api_key})

    print(json.dumps({
        "backup_root": backup_root,
        "tenant_id": tenant["id"],
        "service_id": service["id"],
        "overview": overview,
        "dashboard_count": len(dashboards),
        "archive_stats": replay_stats,
    }, indent=2))


if __name__ == "__main__":
    main()
