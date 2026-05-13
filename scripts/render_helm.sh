#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_PATH="${1:-/tmp/pulselens-helm-rendered.yaml}"

docker run --rm \
  -v "${ROOT_DIR}:/workspace" \
  -w /workspace \
  alpine/helm:3.16.2 \
  template pulselens ./deploy/helm/pulselens > "${OUTPUT_PATH}"

echo "rendered=${OUTPUT_PATH}"

