#!/usr/bin/env python3
import json
import subprocess
import time
import urllib.error
import urllib.request

BASE_TENANT = "http://127.0.0.1:8081"
BASE_QUERY = "http://127.0.0.1:8084"
INTERNAL_TOKEN = "pulselens-internal-token"
POSTGRES_SERVICE = "postgres"


def req(url, method="GET", body=None, headers=None):
    headers = headers or {}
    payload = None
    if body is not None:
        payload = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, method=method, data=payload, headers=headers)
    with urllib.request.urlopen(request, timeout=20) as response:
        return json.loads(response.read().decode()), response.status


def req_status(url, headers=None):
    headers = headers or {}
    try:
        request = urllib.request.Request(url, headers=headers)
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
    email = f"chaos-postgres+{suffix}@pulselens.local"
    tenant, _ = req(
        BASE_TENANT + "/internal/api/v1/tenants",
        "POST",
        {"name": "Chaos PG Tenant", "slug": f"chaos-pg-{suffix}", "plan": "starter", "ingest_quota": 1000, "retention_days": 7, "admin_name": "Chaos", "admin_email": email, "admin_password": "password123"},
        {"X-Internal-Token": INTERNAL_TOKEN},
    )
    token, _ = req(BASE_TENANT + "/api/v1/auth/login", "POST", {"email": email, "password": "password123"})
    return token["data"]["token"], tenant["data"]["tenant"]["id"]


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
    token, tenant_id = bootstrap()
    before = dependency(token, "postgres")
    compose("stop", POSTGRES_SERVICE)
    time.sleep(2)
    try:
        down = wait_status(token, "postgres", "down", timeout=20)
        tenant_ready = req_status("http://127.0.0.1:8081/ready")
        query_ready = req_status("http://127.0.0.1:8084/ready")
        tenant_get = req_status(
            f"{BASE_TENANT}/admin/api/v1/tenants/{tenant_id}",
            headers={"Authorization": f"Bearer {token}"},
        )
    finally:
        compose("start", POSTGRES_SERVICE)
    time.sleep(8)
    recovered = wait_status(token, "postgres", "healthy", timeout=40)
    print(json.dumps({
        "before": before,
        "down": down,
        "tenant_ready": {"status_code": tenant_ready[0], "body": tenant_ready[1]},
        "query_ready": {"status_code": query_ready[0], "body": query_ready[1]},
        "tenant_get": {"status_code": tenant_get[0], "body": tenant_get[1]},
        "recovered": recovered,
    }, indent=2))


if __name__ == "__main__":
    main()
