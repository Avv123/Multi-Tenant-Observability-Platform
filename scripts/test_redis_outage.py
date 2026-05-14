#!/usr/bin/env python3
import json
import subprocess
import time
import urllib.request

BASE_TENANT = "http://127.0.0.1:8081"
BASE_QUERY = "http://127.0.0.1:8084"
INTERNAL_TOKEN = "pulselens-internal-token"
REDIS_SERVICE = "redis"


def req(url, method="GET", body=None, headers=None):
    headers = headers or {}
    payload = None
    if body is not None:
        payload = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, method=method, data=payload, headers=headers)
    with urllib.request.urlopen(request, timeout=20) as response:
        return json.loads(response.read().decode())


def compose(*args):
    subprocess.run(["docker", "compose", *args], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)


def compose_output(*args):
    return subprocess.run(["docker", "compose", *args], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True).stdout.strip()


def bootstrap_token():
    suffix = str(int(time.time()))
    email = f"chaos-redis+{suffix}@pulselens.local"
    req(
        BASE_TENANT + "/internal/api/v1/tenants",
        "POST",
        {
            "name": "Chaos Redis Tenant",
            "slug": f"chaos-redis-{suffix}",
            "plan": "starter",
            "ingest_quota": 1000,
            "retention_days": 7,
            "admin_name": "Chaos",
            "admin_email": email,
            "admin_password": "password123",
        },
        {"X-Internal-Token": INTERNAL_TOKEN},
    )
    return req(BASE_TENANT + "/api/v1/auth/login", "POST", {"email": email, "password": "password123"})["data"]["token"]


def dependency_row(token, name):
    rows = req(BASE_QUERY + "/api/v1/platform/dependencies", headers={"Authorization": f"Bearer {token}"})
    for row in rows["data"]:
        if row["name"] == name:
            return row
    return None


def wait_for_status(token, name, status, timeout=40):
    deadline = time.time() + timeout
    while time.time() < deadline:
        row = dependency_row(token, name)
        if row and row["status"] == status:
            return row
        time.sleep(1)
    raise RuntimeError(f"{name} did not reach {status}")


def wait_for_dependency_row(token, name, timeout=20):
    deadline = time.time() + timeout
    while time.time() < deadline:
        row = dependency_row(token, name)
        if row:
            return row
        time.sleep(1)
    return None


def main():
    token = bootstrap_token()
    before = dependency_row(token, "redis")
    compose("stop", REDIS_SERVICE)
    time.sleep(2)
    down_state = compose_output("ps", REDIS_SERVICE, "--format", "json")
    try:
        down = wait_for_status(token, "redis", "down", timeout=15)
    finally:
        compose("start", REDIS_SERVICE)
    time.sleep(5)
    recovered = wait_for_status(token, "redis", "healthy")
    print(json.dumps({
        "before": before,
        "down_state": down_state,
        "down": down,
        "recovered": recovered,
    }, indent=2))


if __name__ == "__main__":
    main()
