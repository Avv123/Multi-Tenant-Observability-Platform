import json
import os
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request


def req(url, method="GET", body=None, headers=None):
    headers = headers or {}
    payload = None
    if body is not None:
        payload = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, data=payload, method=method, headers=headers)
    with urllib.request.urlopen(request, timeout=15) as response:
        return json.loads(response.read().decode())

def req_expect_status(url, expected_status, method="GET", body=None, headers=None):
    headers = headers or {}
    payload = None
    if body is not None:
        payload = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    request = urllib.request.Request(url, data=payload, method=method, headers=headers)
    try:
        with urllib.request.urlopen(request, timeout=15) as response:
            status_code = response.status
            payload_data = json.loads(response.read().decode())
    except urllib.error.HTTPError as error:
        status_code = error.code
        payload_data = json.loads(error.read().decode())
    if status_code != expected_status:
        raise RuntimeError(f"expected status {expected_status}, got {status_code}: {payload_data}")
    return payload_data

def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]

def wait_until(predicate, timeout_seconds=15, interval_seconds=0.5):
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        value = predicate()
        if value:
            return value
        time.sleep(interval_seconds)
    return None


def mailhog_message_count():
    with urllib.request.urlopen("http://127.0.0.1:8026/api/v2/messages", timeout=10) as response:
        payload = json.loads(response.read().decode())
        return len(payload.get("items", []))


def main():
    internal_token = "pulselens-internal-token"
    base_tenant = "http://127.0.0.1:8081"
    base_ingest = "http://127.0.0.1:8082"
    base_query = "http://127.0.0.1:8084"
    base_ui = "http://127.0.0.1:3000"
    base_alerting = "http://127.0.0.1:8085"
    webhook_port = free_port()
    capture_file = tempfile.NamedTemporaryFile(prefix="pulselens-webhook-", suffix=".jsonl", delete=False)
    capture_file.close()
    webhook_server = subprocess.Popen(
        ["python3", "scripts/test_webhook_receiver.py", "--port", str(webhook_port), "--output", capture_file.name],
    )
    slug = f"smoke-{int(time.time())}"
    email = f"admin+{slug}@acme.local"

    try:
        readiness = {}
        for name, base in {
            "tenant": base_tenant,
            "ingest": base_ingest,
            "query": base_query,
            "alerting": base_alerting,
        }.items():
            readiness[name] = req(base + "/ready")

        create_tenant = req(
            base_tenant + "/internal/api/v1/tenants",
            "POST",
            {
                "name": "Smoke Tenant",
                "slug": slug,
                "plan": "starter",
                "ingest_quota": 1000,
                "retention_days": 7,
                "admin_name": "Smoke Admin",
                "admin_email": email,
                "admin_password": "password123",
            },
            {"X-Internal-Token": internal_token},
        )
        tenant_id = create_tenant["data"]["tenant"]["id"]

        create_service = req(
            base_tenant + f"/internal/api/v1/tenants/{tenant_id}/services",
            "POST",
            {"name": "checkout-api", "environment": "local", "tags": {"team": "platform"}},
            {"X-Internal-Token": internal_token},
        )
        service_id = create_service["data"]["id"]

        login = req(
            base_tenant + "/api/v1/auth/login",
            "POST",
            {"email": email, "password": "password123"},
        )
        token = login["data"]["token"]

        create_key = req(
            base_tenant + "/admin/api/v1/api-keys",
            "POST",
            {
                "tenant_id": tenant_id,
                "service_id": service_id,
                "name": "smoke-key",
                "scopes": ["ingest", "query", "admin"],
            },
            {"Authorization": f"Bearer {token}"},
        )
        api_key = create_key["data"]["key"]
        listed_keys = req(base_tenant + "/admin/api/v1/api-keys", headers={"Authorization": f"Bearer {token}"})
        key_id = listed_keys["data"][0]["id"]
        rotated_key = req(
            base_tenant + f"/admin/api/v1/api-keys/{key_id}/rotate",
            "POST",
            {"name": "smoke-key-rotated"},
            {"Authorization": f"Bearer {token}"},
        )["data"]["key"]
        req_expect_status(
            base_ingest + "/api/v1/ingest",
            403,
            "POST",
            {"events": [{"event_type": "log", "severity": "error", "payload": {"message": "old-key-blocked", "logger": "smoke"}}]},
            {"X-API-Key": api_key},
        )
        req(
            base_ingest + "/api/v1/ingest",
            "POST",
            {"events": [{"event_type": "log", "severity": "info", "payload": {"message": "rotated-key-works", "logger": "smoke"}}]},
            {"X-API-Key": rotated_key},
        )
        api_key = rotated_key
        viewer_email = f"viewer+{slug}@acme.local"
        create_user = req(
            base_tenant + f"/admin/api/v1/tenants/{tenant_id}/users",
            "POST",
            {
                "name": "Viewer User",
                "email": viewer_email,
                "password": "viewer-pass",
                "role": "viewer",
            },
            {"Authorization": f"Bearer {token}"},
        )
        viewer_user_id = create_user["data"]["id"]
        viewer_login = req(
            base_tenant + "/api/v1/auth/login",
            "POST",
            {"email": viewer_email, "password": "viewer-pass"},
        )
        viewer_token = viewer_login["data"]["token"]

        events = [
            {"event_type": "log", "severity": "error", "payload": {"message": "checkout failed for order 42", "logger": "smoke"}},
            {"event_type": "metric", "payload": {"metric_name": "checkout_latency_ms", "value": 148.2, "unit": "ms"}},
            {
                "event_type": "trace",
                "trace_id": "trace-smoke-1",
                "payload": {
                    "span_id": "span-root",
                    "parent_span_id": "",
                    "operation": "checkout",
                    "status": "ok",
                    "start_time": "2025-01-01T00:00:00Z",
                    "end_time": "2025-01-01T00:00:00.150Z",
                },
            },
        ]
        for event in events:
            req(base_ingest + "/api/v1/ingest", "POST", {"events": [event]}, {"X-API-Key": api_key})

        req(
            base_alerting + "/api/v1/notification-channels",
            "POST",
            {
                "name": "Smoke Webhook Channel",
                "type": "webhook",
                "config": {
                    "url": f"http://127.0.0.1:{webhook_port}/webhooks/incidents",
                    "method": "POST",
                    "headers": {"X-Smoke-Test": "true"},
                    "timeout_seconds": 5,
                },
            },
            {"Authorization": f"Bearer {token}"},
        )
        req(
            base_alerting + "/api/v1/notification-channels",
            "POST",
            {
                "name": "Smoke Slack Channel",
                "type": "slack_webhook",
                "config": {
                    "url": f"http://127.0.0.1:{webhook_port}/slack/incidents",
                    "timeout_seconds": 5,
                },
            },
            {"Authorization": f"Bearer {token}"},
        )
        baseline_mailhog_count = mailhog_message_count()
        req(
            base_alerting + "/api/v1/notification-channels",
            "POST",
            {
                "name": "Smoke Email Channel",
                "type": "email",
                "config": {
                    "to": [f"alerts+{slug}@pulselens.local"],
                    "subject_prefix": "[Smoke]",
                },
            },
            {"Authorization": f"Bearer {token}"},
        )

        req(
            base_alerting + "/api/v1/alert-rules",
            "POST",
            {
                "service_id": service_id,
                "name": "Smoke Error Burst",
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
            base_query + "/api/v1/saved-queries",
            "POST",
            {
                "name": "Smoke Errors",
                "query_type": "logs",
                "definition": {"service_id": service_id, "severity": "error", "limit": 20},
            },
            {"Authorization": f"Bearer {token}"},
        )

        req(
            base_query + "/api/v1/dashboards",
            "POST",
            {
                "name": "Smoke Dashboard",
                "description": "smoke dashboard",
                "layout": {"columns": 2},
                "widgets": [{"type": "stat", "metric": "log_count"}],
            },
            {"Authorization": f"Bearer {token}"},
        )

        log_count_before = req(base_query + "/api/v1/overview", headers={"Authorization": f"Bearer {token}"})["data"]["log_count"]
        req(
            base_ingest + "/api/v1/ingest",
            "POST",
            {"events": [{"event_type": "log", "severity": "critical", "payload": {"message": "cache invalidation check", "logger": "smoke"}}]},
            {"X-API-Key": api_key},
        )
        cache_invalidated = wait_until(
            lambda: req(base_query + "/api/v1/overview", headers={"Authorization": f"Bearer {token}"})["data"]["log_count"] > log_count_before,
            timeout_seconds=20,
        )
        if not cache_invalidated:
            raise RuntimeError("overview cache did not invalidate after ingest")

        time.sleep(3)

        webhook_delivery = wait_until(
            lambda: os.path.getsize(capture_file.name) > 0 and open(capture_file.name, "r", encoding="utf-8").read().strip(),
            timeout_seconds=20,
        )
        if not webhook_delivery:
            raise RuntimeError("webhook delivery was not captured")
        webhook_records = [json.loads(line) for line in open(capture_file.name, "r", encoding="utf-8").read().splitlines() if line.strip()]
        if len(webhook_records) < 2:
            raise RuntimeError("expected both webhook and slack deliveries")
        email_delivery = wait_until(
            lambda: mailhog_message_count() > baseline_mailhog_count,
            timeout_seconds=20,
        )
        if not email_delivery:
            raise RuntimeError("email delivery was not captured in MailHog")

        time.sleep(5)

        auth_header = {"Authorization": f"Bearer {token}"}
        viewer_auth_header = {"Authorization": f"Bearer {viewer_token}"}
        overview = req(base_query + "/api/v1/overview", headers=auth_header)
        platform_overview = req(base_query + "/api/v1/platform/overview?limit=10", headers=auth_header)
        platform_dependencies = req(base_query + "/api/v1/platform/dependencies", headers=auth_header)
        platform_kafka_lag = req(base_query + "/api/v1/platform/kafka-lag", headers=auth_header)
        service_health = req(base_query + "/api/v1/services/health", headers=auth_header)
        logs = req(base_query + "/api/v1/logs", headers=auth_header)
        metrics = req(base_query + "/api/v1/metrics", headers=auth_header)
        traces = req(base_query + "/api/v1/traces", headers=auth_header)
        trace_detail = req(base_query + "/api/v1/traces/trace-smoke-1", headers=auth_header)
        log_severity = req(base_query + "/api/v1/analytics/log-severity?limit=20", headers=auth_header)
        metric_series = req(base_query + "/api/v1/analytics/metric-series?limit=20&metric_name=checkout_latency_ms", headers=auth_header)
        trace_latency = req(base_query + "/api/v1/analytics/trace-latency?limit=20", headers=auth_header)
        incidents = req(base_alerting + "/api/v1/incidents", headers=auth_header)
        incident_id = incidents["data"][0]["id"]
        assigned_incident = req(
            base_alerting + f"/api/v1/incidents/{incident_id}/assign",
            "POST",
            {"assigned_to": viewer_user_id},
            auth_header,
        )
        viewer_overview = req(base_query + "/api/v1/overview", headers=viewer_auth_header)
        viewer_forbidden = req_expect_status(
            base_alerting + "/api/v1/alert-rules",
            403,
            "POST",
            {
                "service_id": service_id,
                "name": "Viewer Cannot Create",
                "signal_type": "log",
                "severity": "error",
                "aggregation": "count",
                "comparator": ">=",
                "threshold": 1,
                "window_minutes": 5,
                "cooldown_minutes": 1,
            },
            viewer_auth_header,
        )
        saved_queries = req(base_query + "/api/v1/saved-queries", headers=auth_header)
        dashboards = req(base_query + "/api/v1/dashboards", headers=auth_header)
        notification_channels = req(base_alerting + "/api/v1/notification-channels", headers=auth_header)
        notification_deliveries = req(base_alerting + "/api/v1/notification-deliveries", headers=auth_header)
        audit_logs = req(base_tenant + f"/admin/api/v1/tenants/{tenant_id}/audit-logs", headers=auth_header)
        users = req(base_tenant + f"/admin/api/v1/tenants/{tenant_id}/users", headers=auth_header)
        with urllib.request.urlopen(base_ui + "/", timeout=15) as response:
            index_html = response.read().decode()

        print(
            json.dumps(
                {
                    "readiness": readiness,
                    "tenant_id": tenant_id,
                    "service_id": service_id,
                    "api_key_prefix": api_key[:12],
                    "overview": overview["data"],
                    "platform_runtime_count": len(platform_overview["data"]["runtime"]),
                    "platform_backpressure_count": len(platform_overview["data"]["backpressure"]),
                    "cleanup_run_count": len(platform_overview["data"]["cleanup_runs"]),
                    "dependency_count": len(platform_dependencies["data"]),
                    "kafka_lag_rows": len(platform_kafka_lag["data"]),
                    "service_health_count": len(service_health["data"]),
                    "log_count": len(logs["data"]),
                    "metric_count": len(metrics["data"]),
                    "trace_count": len(traces["data"]),
                    "trace_detail_count": len(trace_detail["data"]),
                    "log_severity_count": len(log_severity["data"]),
                    "metric_series_count": len(metric_series["data"]),
                    "trace_latency_count": len(trace_latency["data"]),
                    "incident_count": len(incidents["data"]),
                    "assigned_incident_user": assigned_incident["data"]["assigned_to"],
                    "saved_query_count": len(saved_queries["data"]),
                    "dashboard_count": len(dashboards["data"]),
                    "notification_channel_count": len(notification_channels["data"]),
                    "notification_delivery_count": len(notification_deliveries["data"]),
                    "audit_log_count": len(audit_logs["data"]),
                    "user_count": len(users["data"]),
                    "viewer_overview_role": viewer_overview["data"]["tenant_id"],
                    "viewer_forbidden_code": viewer_forbidden["error_code"],
                    "webhook_delivery_count": len(webhook_records),
                    "webhook_first_event": webhook_records[0]["body"]["event_type"],
                    "mailhog_message_count": mailhog_message_count(),
                    "ui_served": "PulseLens" in index_html,
                },
                indent=2,
            )
        )
    finally:
        webhook_server.terminate()
        webhook_server.wait(timeout=5)
        if os.path.exists(capture_file.name):
            os.remove(capture_file.name)


if __name__ == "__main__":
    main()
