# Chaos Drills

## Available
- `python3 scripts/test_redis_outage.py`
- `python3 scripts/test_clickhouse_outage.py`
- `python3 scripts/test_kafka_outage.py`
- `python3 scripts/test_postgres_outage.py`
- `python3 scripts/test_minio_outage.py`
- `python3 scripts/test_multi_dependency_outage.py`

## Expected shape
Each drill reports:
- dependency health before outage
- degraded state during outage
- recovered state after restart
- affected API/readiness behavior
