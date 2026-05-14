#!/usr/bin/env bash
set -euo pipefail

CLICKHOUSE_URL="${CLICKHOUSE_URL:-http://127.0.0.1:8123}"
CLICKHOUSE_DB="${CLICKHOUSE_DB:-pulselens}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-omniful}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-omniful}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

docker compose exec -T postgres psql -U omniful -d pulselens -v ON_ERROR_STOP=1 -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

python3 - "${CLICKHOUSE_URL}" "${CLICKHOUSE_DB}" "${CLICKHOUSE_USER}" "${CLICKHOUSE_PASSWORD}" <<'PY'
import json
import sys
import urllib.parse
import urllib.request

base = sys.argv[1].rstrip("/")
database = sys.argv[2]
username = sys.argv[3]
password = sys.argv[4]

def request(query):
    params = urllib.parse.urlencode({"database": database, "query": query})
    req = urllib.request.Request(f"{base}/?{params}", method="POST")
    if username:
        import base64
        req.add_header("Authorization", "Basic " + base64.b64encode(f"{username}:{password}".encode()).decode())
    with urllib.request.urlopen(req, timeout=60) as response:
        return response.read().decode()

tables = json.loads(request("SHOW TABLES FORMAT JSON"))["data"]
for row in tables:
    request(f"TRUNCATE TABLE {row['name']}")
PY

docker compose exec -T minio sh -lc 'rm -rf /data/* && mkdir -p /data'
