#!/usr/bin/env python3
import json
import subprocess
import time
import urllib.error
import urllib.request

BASE_TENANT = "http://127.0.0.1:8081"
BASE_INGEST = "http://127.0.0.1:8082"
BASE_QUERY = "http://127.0.0.1:8084"
BASE_ARCHIVE = "http://127.0.0.1:8086"
INTERNAL_TOKEN = "pulselens-internal-token"
MINIO_SERVICE = "minio"


def req(url, method="GET", body=None, headers=None):
    headers = headers or {}
    payload = None
    if body is not None:
        payload = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, method=method, data=payload, headers=headers)
    with urllib.request.urlopen(request, timeout=20) as response:
        return json.loads(response.read().decode()), response.status


def req_status(url, method="GET", body=None, headers=None):
    headers = headers or {}
    payload = None
    if body is not None:
        payload = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, method=method, data=payload, headers=headers)
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            return response.status, response.read().decode()
    except urllib.error.HTTPError as error:
        return error.code, error.read().decode()
    except Exception as error:
        return 0, str(error)


def compose(*args):
    subprocess.run(["docker", "compose", *args], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)


def bootstrap():
    suffix = str(int(time.time()))
    email = f"chaos-minio+{suffix}@pulselens.local"
    tenant, _ = req(
        BASE_TENANT + "/internal/api/v1/tenants",
        "POST",
        {"name": "Chaos MinIO Tenant", "slug": f"chaos-minio-{suffix}", "plan": "starter", "ingest_quota": 1000, "retention_days": 7, "admin_name": "Chaos", "admin_email": email, "admin_password": "password123"},
        {"X-Internal-Token": INTERNAL_TOKEN},
    )
    token, _ = req(BASE_TENANT + "/api/v1/auth/login", "POST", {"email": email, "password": "password123"})
    tenant_id = tenant["data"]["tenant"]["id"]
    service, _ = req(
        BASE_TENANT + f"/admin/api/v1/tenants/{tenant_id}/services",
        "POST",
        {"name": "chaos-minio-api", "environment": "local"},
        {"Authorization": f"Bearer {token['data']['token']}"},
    )
    key, _ = req(
        BASE_TENANT + "/admin/api/v1/api-keys",
        "POST",
        {"tenant_id": tenant_id, "service_id": service["data"]["id"], "name": "chaos-minio-key", "scopes": ["ingest", "query", "admin"]},
        {"Authorization": f"Bearer {token['data']['token']}"},
    )
    return token["data"]["token"], key["data"]["key"], service["data"]["id"]


def dependency(token, name):
    rows, _ = req(BASE_QUERY + "/api/v1/platform/dependencies", headers={"Authorization": f"Bearer {token}"})
    for row in rows["data"]:
        if row["name"] == name:
            return row
    return None


def wait_status(token, name, status, timeout=30):
    deadline = time.time() + timeout
    while time.time() < deadline:
        row = dependency(token, name)
        if row and row["status"] == status:
            return row
        time.sleep(1)
    raise RuntimeError(f"{name} did not reach {status}")


def main():
    token, api_key, service_id = bootstrap()
    before = dependency(token, "minio")
    compose("stop", MINIO_SERVICE)
    time.sleep(2)
    try:
        down = wait_status(token, "minio", "down", timeout=20)
        archive_ready = req_status("http://127.0.0.1:8086/ready")
        req_status(
            BASE_INGEST + "/api/v1/ingest",
            "POST",
            {"events": [{"event_type": "log", "severity": "error", "payload": {"message": "chaos-minio", "logger": "chaos"}}]},
            {"X-API-Key": api_key},
        )
        time.sleep(4)
        replay_create = req_status(
            BASE_ARCHIVE + "/api/v1/replay-jobs",
            "POST",
            {"service_id": service_id, "event_type": "log", "start_time": "2024-12-31T23:59:00Z", "end_time": "2025-01-01T00:10:00Z"},
            {"Authorization": f"Bearer {token}"},
        )
    finally:
        compose("start", MINIO_SERVICE)
    time.sleep(8)
    recovered = wait_status(token, "minio", "healthy", timeout=40)
    print(json.dumps({
        "before": before,
        "down": down,
        "archive_ready": {"status_code": archive_ready[0], "body": archive_ready[1]},
        "replay_create": {"status_code": replay_create[0], "body": replay_create[1]},
        "recovered": recovered,
    }, indent=2))


if __name__ == "__main__":
    main()
