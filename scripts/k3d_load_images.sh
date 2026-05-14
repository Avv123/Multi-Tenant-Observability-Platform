#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLUSTER_NAME="${1:-pulselens-local}"

cd "${ROOT_DIR}"
docker compose build tenant-service ingest-service processing-service query-service alerting-service archive-service

for image in \
  pulselens-tenant-service:latest \
  pulselens-ingest-service:latest \
  pulselens-processing-service:latest \
  pulselens-query-service:latest \
  pulselens-alerting-service:latest \
  pulselens-archive-service:latest
do
  k3d image import -c "${CLUSTER_NAME}" "${image}"
done
