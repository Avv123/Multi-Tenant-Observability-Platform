#!/usr/bin/env python3
import json
import statistics
import threading
import time
import urllib.request


BASE_TENANT = "http://127.0.0.1:8081"
BASE_INGEST = "http://127.0.0.1:8082"
BASE_QUERY = "http://127.0.0.1:8084"
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


def bootstrap():
    suffix = str(int(time.time()))
    email = f"soak-admin+{suffix}@pulselens.local"
    tenant = req(
        BASE_TENANT + "/internal/api/v1/tenants",
        "POST",
        {
            "name": "Soak Tenant",
            "slug": f"soak-{suffix}",
            "plan": "starter",
            "ingest_quota": 10000,
            "retention_days": 7,
            "admin_name": "Soak Admin",
            "admin_email": email,
            "admin_password": "password123",
        },
        {"X-Internal-Token": INTERNAL_TOKEN},
    )["data"]["tenant"]
    login = req(BASE_TENANT + "/api/v1/auth/login", "POST", {"email": email, "password": "password123"})
    token = login["data"]["token"]
    service = req(
        BASE_TENANT + f"/admin/api/v1/tenants/{tenant['id']}/services",
        "POST",
        {"name": "soak-api", "environment": "local", "tags": {"source": "soak"}},
        {"Authorization": f"Bearer {token}"},
    )["data"]
    api_key = req(
        BASE_TENANT + "/admin/api/v1/api-keys",
        "POST",
        {"tenant_id": tenant["id"], "service_id": service["id"], "name": "soak-key", "scopes": ["ingest", "query", "admin"]},
        {"Authorization": f"Bearer {token}"},
    )["data"]["key"]
    return token, api_key


def main():
    duration_seconds = 20
    token, api_key = bootstrap()
    stop_at = time.time() + duration_seconds
    ingest_latencies = []
    query_latencies = []
    failures = []

    def ingest_loop():
        while time.time() < stop_at:
            started = time.time()
            try:
                req(
                    BASE_INGEST + "/api/v1/ingest",
                    "POST",
                    {
                        "events": [
                            {"event_type": "log", "severity": "error", "payload": {"message": f"soak-log-{time.time()}", "logger": "soak"}},
                            {"event_type": "metric", "payload": {"metric_name": "soak_latency_ms", "value": 120 + int(time.time()) % 10}},
                            {
                                "event_type": "trace",
                                "trace_id": f"soak-trace-{int(time.time() * 1000)}",
                                "payload": {
                                    "span_id": "root",
                                    "parent_span_id": "",
                                    "operation": "checkout",
                                    "status": "ok",
                                    "start_time": "2025-01-01T00:00:00Z",
                                    "end_time": "2025-01-01T00:00:00.120Z",
                                },
                            },
                        ]
                    },
                    {"X-API-Key": api_key},
                )
                ingest_latencies.append((time.time() - started) * 1000)
            except Exception as error:
                failures.append(f"ingest:{error}")
            time.sleep(0.15)

    def query_loop():
        while time.time() < stop_at:
            started = time.time()
            try:
                req(BASE_QUERY + "/api/v1/overview", headers={"Authorization": f"Bearer {token}"})
                req(BASE_QUERY + "/api/v1/analytics/metric-series?metric_name=soak_latency_ms&limit=10", headers={"Authorization": f"Bearer {token}"})
                query_latencies.append((time.time() - started) * 1000)
            except Exception as error:
                failures.append(f"query:{error}")
            time.sleep(0.2)

    ingest_thread = threading.Thread(target=ingest_loop)
    query_thread = threading.Thread(target=query_loop)
    ingest_thread.start()
    query_thread.start()
    ingest_thread.join()
    query_thread.join()

    print(json.dumps({
        "duration_seconds": duration_seconds,
        "ingest_requests": len(ingest_latencies),
        "query_requests": len(query_latencies),
        "failures": failures,
        "ingest_p95_ms": round(statistics.quantiles(ingest_latencies, n=20)[18], 2) if len(ingest_latencies) >= 20 else round(max(ingest_latencies or [0]), 2),
        "query_p95_ms": round(statistics.quantiles(query_latencies, n=20)[18], 2) if len(query_latencies) >= 20 else round(max(query_latencies or [0]), 2),
    }, indent=2))


if __name__ == "__main__":
    main()
