#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${ROOT_DIR}"

port_busy() {
  local port="$1"
  lsof -nP -iTCP:"${port}" -sTCP:LISTEN >/dev/null 2>&1
}

wait_for() {
  local description="$1"
  local command="$2"
  local retries="${3:-60}"
  for _ in $(seq 1 "${retries}"); do
    if eval "${command}" >/dev/null 2>&1; then
      echo "ready: ${description}"
      return 0
    fi
    sleep 2
  done
  echo "timeout waiting for ${description}" >&2
  return 1
}

postgres_ready_command() {
  if command -v psql >/dev/null 2>&1; then
    echo "PGPASSWORD=omniful psql -h 127.0.0.1 -U omniful -d postgres -c 'select 1' >/dev/null"
    return
  fi
  if docker ps --format '{{.Names}}' | grep -q '^postgres-db$'; then
    echo "docker exec postgres-db psql -U omniful -d postgres -c 'select 1' >/dev/null"
    return
  fi
  if docker ps --format '{{.Names}}' | grep -q '^pulselens-postgres$'; then
    echo "docker exec pulselens-postgres psql -U omniful -d postgres -c 'select 1' >/dev/null"
    return
  fi
  echo "false"
}

for container in pulselens-redis pulselens-clickhouse pulselens-minio pulselens-postgres pulselens-kafka pulselens-mailhog; do
  if docker ps -a --format '{{.Names}}' | grep -q "^${container}\$"; then
    docker rm -f "${container}" >/dev/null 2>&1 || true
  fi
done

compose_services=()

if ! port_busy 6381; then
  compose_services+=("redis")
else
  echo "using existing redis on localhost:6381"
fi

if ! port_busy 8123; then
  compose_services+=("clickhouse")
else
  echo "using existing clickhouse on localhost:8123"
fi

if ! port_busy 9010; then
  compose_services+=("minio")
else
  echo "using existing minio-compatible store on localhost:9010"
fi

if ! port_busy 8025; then
  compose_services+=("mailhog")
else
  echo "using existing mailhog on localhost:8025"
fi

if ! port_busy 5432; then
  compose_services+=("postgres")
else
  echo "using existing postgres on localhost:5432"
fi

if ! port_busy 9092; then
  compose_services+=("kafka")
  start_kafka_init=true
else
  echo "using existing kafka on localhost:9092"
  start_kafka_init=false
fi

if [[ "${#compose_services[@]}" -gt 0 ]]; then
  docker compose up -d "${compose_services[@]}"
fi

if [[ "${start_kafka_init}" == true ]]; then
  docker compose up -d kafka-init
fi

wait_for "postgres" "$(postgres_ready_command)"
wait_for "redis" "redis-cli -h 127.0.0.1 -p 6381 ping | grep -q PONG"
wait_for "kafka" "bash -lc 'echo > /dev/tcp/127.0.0.1/9092'"
wait_for "clickhouse" "curl -fsS http://127.0.0.1:8123/ping | grep -q Ok."
wait_for "minio" "curl -fsS http://127.0.0.1:9010/minio/health/live"
wait_for "mailhog" "curl -fsS http://127.0.0.1:8025/api/v2/messages"

if docker ps --format '{{.Names}}' | grep -q '^kafka$'; then
  docker exec kafka bash -lc 'for topic in pulselens.logs.v1 pulselens.metrics.v1 pulselens.traces.v1 pulselens.custom.v1 pulselens.retry.v1 pulselens.retry.scheduled.v1; do kafka-topics --bootstrap-server localhost:9092 --create --if-not-exists --topic "$topic" --partitions 1 --replication-factor 1; done'
elif docker ps --format '{{.Names}}' | grep -q '^pulselens-kafka$'; then
  docker exec pulselens-kafka rpk topic create pulselens.logs.v1 --brokers localhost:9092 >/dev/null 2>&1 || true
  docker exec pulselens-kafka rpk topic create pulselens.metrics.v1 --brokers localhost:9092 >/dev/null 2>&1 || true
  docker exec pulselens-kafka rpk topic create pulselens.traces.v1 --brokers localhost:9092 >/dev/null 2>&1 || true
  docker exec pulselens-kafka rpk topic create pulselens.custom.v1 --brokers localhost:9092 >/dev/null 2>&1 || true
  docker exec pulselens-kafka rpk topic create pulselens.retry.v1 --brokers localhost:9092 >/dev/null 2>&1 || true
  docker exec pulselens-kafka rpk topic create pulselens.retry.scheduled.v1 --brokers localhost:9092 >/dev/null 2>&1 || true
fi

mkdir -p data/archive

echo "infra ready"
