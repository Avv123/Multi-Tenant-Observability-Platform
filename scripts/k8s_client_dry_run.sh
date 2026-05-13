#!/usr/bin/env bash
set -euo pipefail

RENDERED_PATH="${1:-/tmp/pulselens-helm-rendered.yaml}"

if [[ ! -f "${RENDERED_PATH}" ]]; then
  echo "rendered manifest not found: ${RENDERED_PATH}" >&2
  exit 1
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "skipped: kubectl not installed"
  exit 0
fi

kubectl apply --dry-run=client --validate=false -f "${RENDERED_PATH}" >/dev/null
echo "dry-run=client succeeded for ${RENDERED_PATH}"
