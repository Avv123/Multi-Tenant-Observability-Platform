#!/usr/bin/env python3
import json
import subprocess
import time
import urllib.error
import urllib.request

BASE_TENANT = "http://127.0.0.1:8081"
BASE_INGEST = "http://127.0.0.1:8082"
BASE_QUERY = "http://127.0.0.1:8084"
INTERNAL_TOKEN = "pulselens-internal-token"
KAFKA_SERVICE = "kafka"


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
    email = f"chaos-kafka+{suffix}@pulselens.local"
    tenant, _ = req(
        BASE_TENANT + "/internal/api/v1/tenants",
        "POST",
        {"name": "Chaos Kafka Tenant", "slug": f"chaos-kafka-{suffix}", "plan": "starter", "ingest_quota": 1000, "retention_days": 7, "admin_name": "Chaos", "admin_email": email, "admin_password": "password123"},
        {"X-Internal-Token": INTERNAL_TOKEN},
    )
    token, _ = req(BASE_TENANT + "/api/v1/auth/login", "POST", {"email": email, "password": "password123"})
    tenant_id = tenant["data"]["tenant"]["id"]
    service, _ = req(
        BASE_TENANT + f"/admin/api/v1/tenants/{tenant_id}/services",
        "POST",
        {"name": "chaos-kafka-api", "environment": "local"},
        {"Authorization": f"Bearer {token['data']['token']}"},
    )
    key, _ = req(
        BASE_TENANT + "/admin/api/v1/api-keys",
        "POST",
        {"tenant_id": tenant_id, "service_id": service["data"]["id"], "name": "chaos-kafka-key", "scopes": ["ingest", "query", "admin"]},
        {"Authorization": f"Bearer {token['data']['token']}"},
    )
    return token["data"]["token"], key["data"]["key"]


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
    token, api_key = bootstrap()
    before = dependency(token, "kafka")
    before_ready_status, before_ready_body = req_status("http://127.0.0.1:8082/ready")
    compose("stop", KAFKA_SERVICE)
    time.sleep(2)
    try:
        down = wait_status(token, "kafka", "down", timeout=20)
        ingest_ready_status, ingest_ready_body = req_status("http://127.0.0.1:8082/ready")
        ingest_status, ingest_body = req_status(
            BASE_INGEST + "/api/v1/ingest",
            "POST",
            {"events": [{"event_type": "log", "severity": "error", "payload": {"message": "chaos-kafka", "logger": "chaos"}}]},
            {"X-API-Key": api_key},
        )
    finally:
        compose("start", KAFKA_SERVICE)
    time.sleep(6)
    recovered = wait_status(token, "kafka", "healthy", timeout=40)
    after_ready_status, after_ready_body = req_status("http://127.0.0.1:8082/ready")
    print(json.dumps({
        "before": before,
        "before_ready": {"status_code": before_ready_status, "body": before_ready_body},
        "down": down,
        "during_ready": {"status_code": ingest_ready_status, "body": ingest_ready_body},
        "ingest_attempt": {"status_code": ingest_status, "body": ingest_body},
        "recovered": recovered,
        "after_ready": {"status_code": after_ready_status, "body": after_ready_body},
    }, indent=2))


if __name__ == "__main__":
    main()
