#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_ROOT="${1:-${ROOT_DIR}/data/backups/$(date -u +%Y%m%dT%H%M%SZ)}"
TARGET_DIR="${BACKUP_ROOT}/clickhouse"
mkdir -p "${TARGET_DIR}"

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

def request(query):
    params = urllib.parse.urlencode({
        "database": database,
        "output_format_json_quote_64bit_integers": "0",
        "query": query,
    })
    req = urllib.request.Request(f"{base}/?{params}", method="POST")
    if username:
        import base64
        token = base64.b64encode(f"{username}:{password}".encode()).decode()
        req.add_header("Authorization", f"Basic {token}")
    with urllib.request.urlopen(req, timeout=30) as response:
        return response.read().decode()

tables = json.loads(request("SHOW TABLES FORMAT JSON"))["data"]
names = [row["name"] for row in tables if not row["name"].startswith(".")]
for name in names:
    (target / f"{name}.jsonl").write_text(request(f"SELECT * FROM {name} FORMAT JSONEachRow"))

(target / "metadata.json").write_text(json.dumps({
    "database": database,
    "created_at": __import__("datetime").datetime.utcnow().replace(microsecond=0).isoformat() + "Z",
    "tables": names,
}, indent=2))
print(target)
PY
