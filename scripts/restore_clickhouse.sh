#!/usr/bin/env bash
set -euo pipefail

BACKUP_ROOT="${1:?backup root required}"
TARGET_DIR="${BACKUP_ROOT}/clickhouse"
CLICKHOUSE_URL="${CLICKHOUSE_URL:-http://127.0.0.1:8123}"
CLICKHOUSE_DB="${CLICKHOUSE_DB:-pulselens}"
CLICKHOUSE_USER="${CLICKHOUSE_USER:-omniful}"
CLICKHOUSE_PASSWORD="${CLICKHOUSE_PASSWORD:-omniful}"

python3 - "${TARGET_DIR}" "${CLICKHOUSE_URL}" "${CLICKHOUSE_DB}" "${CLICKHOUSE_USER}" "${CLICKHOUSE_PASSWORD}" <<'PY'
import json
import pathlib
import sys
import urllib.parse
import urllib.request

target = pathlib.Path(sys.argv[1])
base = sys.argv[2].rstrip("/")
database = sys.argv[3]
username = sys.argv[4]
password = sys.argv[5]

def request(query, body=None):
    params = urllib.parse.urlencode({
        "database": database,
        "output_format_json_quote_64bit_integers": "0",
        "query": query,
    })
    data = None if body is None else body.encode()
    req = urllib.request.Request(f"{base}/?{params}", data=data, method="POST")
    if username:
        import base64
        token = base64.b64encode(f"{username}:{password}".encode()).decode()
        req.add_header("Authorization", f"Basic {token}")
    with urllib.request.urlopen(req, timeout=60) as response:
        return response.read().decode()

metadata = json.loads((target / "metadata.json").read_text())
tables = metadata["tables"]
for name in tables:
    request(f"TRUNCATE TABLE {name}")
    payload_path = target / f"{name}.jsonl"
    payload = payload_path.read_text() if payload_path.exists() else ""
    if payload.strip():
        request(f"INSERT INTO {name} FORMAT JSONEachRow", payload)
print(target)
PY
