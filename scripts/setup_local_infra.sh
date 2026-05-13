#!/usr/bin/env bash
set -euo pipefail

ocean infra start

if ! docker ps -a --format '{{.Names}}' | grep -q '^pulselens-redis$'; then
  docker run -d \
    --name pulselens-redis \
    -p 6381:6379 \
    redis:7-alpine >/dev/null
else
  docker start pulselens-redis >/dev/null
fi

for _ in $(seq 1 30); do
  if redis-cli -h 127.0.0.1 -p 6381 ping >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

if ! docker ps -a --format '{{.Names}}' | grep -q '^pulselens-clickhouse$'; then
  docker run -d \
    --name pulselens-clickhouse \
    -p 8123:8123 \
    -p 9000:9000 \
    -e CLICKHOUSE_DB=pulselens \
    -e CLICKHOUSE_USER=omniful \
    -e CLICKHOUSE_PASSWORD=omniful \
    -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 \
    clickhouse/clickhouse-server:24.8 >/dev/null
else
  docker start pulselens-clickhouse >/dev/null
fi

for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:8123/ping >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

docker exec postgres-db psql -U omniful -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='pulselens'" | grep -q 1 \
  || docker exec postgres-db psql -U omniful -d postgres -c "CREATE DATABASE pulselens;"

docker exec kafka bash -lc 'for topic in pulselens.logs.v1 pulselens.metrics.v1 pulselens.traces.v1 pulselens.custom.v1 pulselens.retry.v1 pulselens.retry.scheduled.v1; do kafka-topics --bootstrap-server localhost:9092 --create --if-not-exists --topic "$topic" --partitions 1 --replication-factor 1; done'

mkdir -p data/archive
