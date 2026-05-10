#!/usr/bin/env python3
import json
import time
import urllib.request


BASE_TENANT = "http://127.0.0.1:8081"
BASE_INGEST = "http://127.0.0.1:8082"
BASE_ALERTING = "http://127.0.0.1:8085"
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


def main():
    suffix = str(int(time.time()))
    email = f"failure-admin+{suffix}@pulselens.local"
    tenant = req(
        BASE_TENANT + "/internal/api/v1/tenants",
        "POST",
        {
            "name": "Failure Drill Tenant",
            "slug": f"failure-{suffix}",
            "plan": "starter",
            "ingest_quota": 1000,
            "retention_days": 7,
            "admin_name": "Failure Admin",
            "admin_email": email,
            "admin_password": "password123",
        },
        {"X-Internal-Token": INTERNAL_TOKEN},
    )["data"]["tenant"]
    token = req(BASE_TENANT + "/api/v1/auth/login", "POST", {"email": email, "password": "password123"})["data"]["token"]
    service = req(
        BASE_TENANT + f"/admin/api/v1/tenants/{tenant['id']}/services",
        "POST",
        {"name": "failure-api", "environment": "local"},
        {"Authorization": f"Bearer {token}"},
    )["data"]
    api_key = req(
        BASE_TENANT + "/admin/api/v1/api-keys",
        "POST",
        {"tenant_id": tenant["id"], "service_id": service["id"], "name": "failure-key", "scopes": ["ingest", "query", "admin"]},
        {"Authorization": f"Bearer {token}"},
    )["data"]["key"]

    req(
        BASE_ALERTING + "/api/v1/notification-channels",
        "POST",
        {
            "name": "Broken Webhook",
            "type": "webhook",
            "config": {"url": "http://127.0.0.1:65535/unreachable", "method": "POST", "timeout_seconds": 1},
        },
        {"Authorization": f"Bearer {token}"},
    )
    req(
        BASE_ALERTING + "/api/v1/alert-rules",
        "POST",
        {
            "service_id": service["id"],
            "name": "Failure Drill Error Burst",
            "signal_type": "log",
            "severity": "error",
            "aggregation": "count",
            "comparator": ">=",
            "threshold": 1,
            "window_minutes": 10,
            "cooldown_minutes": 1,
        },
        {"Authorization": f"Bearer {token}"},
    )
    req(
        BASE_INGEST + "/api/v1/ingest",
        "POST",
        {"events": [{"event_type": "log", "severity": "error", "payload": {"message": "failure-drill", "logger": "failure"}}]},
        {"X-API-Key": api_key},
    )

    deliveries = []
    for _ in range(30):
        deliveries = req(BASE_ALERTING + "/api/v1/notification-deliveries", headers={"Authorization": f"Bearer {token}"})["data"]
        if deliveries:
            break
        time.sleep(1)

    failed = [delivery for delivery in deliveries if delivery["status"] == "failed"]
    print(json.dumps({
        "delivery_count": len(deliveries),
        "failed_deliveries": len(failed),
        "responses": [delivery.get("response") for delivery in failed[:3]],
    }, indent=2))


if __name__ == "__main__":
    main()
