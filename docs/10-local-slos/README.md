# Local SLOs

These are local quality gates for the Compose-first runtime.

## Thresholds
- ingest `p95` under `500ms` in `scripts/test_load.py`
- query `p95` under `500ms` in `scripts/test_load.py`
- ingest `p95` under `750ms` in `scripts/test_sustained_load.py`
- query `p95` under `750ms` in `scripts/test_sustained_load.py`
- functional failures: `0`
- dependency recovery target: `40s`

## Commands
- `python3 scripts/test_load.py --json-out data/benchmark-reports/load.json`
- `python3 scripts/test_sustained_load.py --json-out data/benchmark-reports/soak.json`
- `python3 scripts/generate_benchmark_report.py data/benchmark-reports data/benchmark-reports/load.json data/benchmark-reports/soak.json`
- `bash scripts/run_full_validation.sh`
