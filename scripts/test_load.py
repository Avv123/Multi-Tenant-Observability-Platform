import json
import argparse
import statistics
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed


def req(url, method="GET", body=None, headers=None):
    headers = headers or {}
    payload = None
    if body is not None:
        payload = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, data=payload, method=method, headers=headers)
    started = time.perf_counter()
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            payload_data = json.loads(response.read().decode())
            status_code = response.status
    except urllib.error.HTTPError as error:
        payload_data = json.loads(error.read().decode())
        status_code = error.code
    elapsed_ms = (time.perf_counter() - started) * 1000
    return payload_data, elapsed_ms, status_code


def percentile(values, ratio):
    if not values:
        return 0
    ordered = sorted(values)
    index = min(len(ordered) - 1, max(0, int(len(ordered) * ratio) - 1))
    return ordered[index]


def bootstrap():
    suffix = int(time.time())
    email = f"load-admin+{suffix}@pulselens.local"
    tenant = req(
        "http://127.0.0.1:8081/internal/api/v1/tenants",
        "POST",
        {
            "name": "Load Tenant",
            "slug": f"load-{suffix}",
            "plan": "starter",
            "ingest_quota": 100000,
            "retention_days": 7,
            "admin_name": "Load Admin",
            "admin_email": email,
            "admin_password": "password123",
        },
        {"X-Internal-Token": "pulselens-internal-token"},
    )[0]
    tenant_id = tenant["data"]["tenant"]["id"]
    login = req(
        "http://127.0.0.1:8081/api/v1/auth/login",
        "POST",
        {"email": email, "password": "password123"},
    )[0]
    token = login["data"]["token"]
    service = req(
        f"http://127.0.0.1:8081/admin/api/v1/tenants/{tenant_id}/services",
        "POST",
        {"name": "load-api", "environment": "local", "tags": {"suite": "load"}},
        {"Authorization": f"Bearer {token}"},
    )[0]
    service_id = service["data"]["id"]
    key = req(
        "http://127.0.0.1:8081/admin/api/v1/api-keys",
        "POST",
        {"tenant_id": tenant_id, "service_id": service_id, "name": "load-key", "scopes": ["ingest", "query", "admin"]},
        {"Authorization": f"Bearer {token}"},
    )[0]
    return tenant_id, service_id, token, key["data"]["key"]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--max-ingest-p95-ms", type=float, default=500)
    parser.add_argument("--max-query-p95-ms", type=float, default=500)
    parser.add_argument("--max-failures", type=int, default=0)
    parser.add_argument("--json-out", type=str, default="")
    args = parser.parse_args()

    _, service_id, token, api_key = bootstrap()
    ingest_latencies = []
    query_latencies = []
    failures = []

    def ingest_task(index):
        event = {
            "events": [
                {
                    "event_type": "log",
                    "severity": "error" if index % 5 == 0 else "info",
                    "payload": {"message": f"load event {index}", "logger": "load-test"},
                }
            ]
        }
        payload, latency, status_code = req("http://127.0.0.1:8082/api/v1/ingest", "POST", event, {"X-API-Key": api_key})
        return latency, status_code, payload

    def query_task():
        payload, latency, status_code = req(
            f"http://127.0.0.1:8084/api/v1/logs?service_id={service_id}&limit=20",
            headers={"Authorization": f"Bearer {token}"},
        )
        return latency, status_code, payload

    started = time.perf_counter()
    with ThreadPoolExecutor(max_workers=16) as executor:
        futures = []
        for index in range(120):
            futures.append(("ingest", executor.submit(ingest_task, index)))
            if index % 3 == 0:
                futures.append(("query", executor.submit(query_task)))

        for kind, future in futures:
            try:
                latency, status_code, payload = future.result()
                if status_code >= 400:
                    failures.append({"kind": kind, "status": status_code, "message": payload.get("message")})
                    continue
                if kind == "ingest":
                    ingest_latencies.append(latency)
                else:
                    query_latencies.append(latency)
            except Exception as error:
                failures.append({"kind": kind, "status": "exception", "message": str(error)})

    duration_seconds = time.perf_counter() - started
    result = {
        "duration_seconds": round(duration_seconds, 2),
        "ingest_requests": len(ingest_latencies),
        "query_requests": len(query_latencies),
        "failures": failures,
        "ingest_p50_ms": round(statistics.median(ingest_latencies), 2) if ingest_latencies else 0,
        "ingest_p95_ms": round(percentile(ingest_latencies, 0.95), 2),
        "query_p50_ms": round(statistics.median(query_latencies), 2) if query_latencies else 0,
        "query_p95_ms": round(percentile(query_latencies, 0.95), 2),
    }
    if args.json_out:
        with open(args.json_out, "w", encoding="utf-8") as handle:
            json.dump(result, handle, indent=2)
    print(json.dumps(result, indent=2))
    if (
        len(failures) > args.max_failures
        or result["ingest_p95_ms"] > args.max_ingest_p95_ms
        or result["query_p95_ms"] > args.max_query_p95_ms
    ):
        raise SystemExit(1)


if __name__ == "__main__":
    main()
