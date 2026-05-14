#!/usr/bin/env bash
set -euo pipefail

RENDERED_PATH="${1:-/tmp/pulselens-helm-rendered.yaml}"

if [[ ! -f "${RENDERED_PATH}" ]]; then
  echo "rendered manifest not found: ${RENDERED_PATH}" >&2
  exit 1
fi

if [[ ! -s "${RENDERED_PATH}" ]]; then
  echo "rendered manifest is empty: ${RENDERED_PATH}" >&2
  exit 1
fi

document_count="$(grep -c '^---' "${RENDERED_PATH}" || true)"
kind_count="$(grep -c '^kind:' "${RENDERED_PATH}" || true)"
api_version_count="$(grep -c '^apiVersion:' "${RENDERED_PATH}" || true)"

if [[ "${kind_count}" -eq 0 || "${api_version_count}" -eq 0 ]]; then
  echo "rendered manifest failed structural sanity checks: missing kind/apiVersion" >&2
  exit 1
fi

echo "structural-sanity succeeded for ${RENDERED_PATH} docs=${document_count} kinds=${kind_count}"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "skipped: kubectl not installed"
  exit 0
fi

server="$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' 2>/dev/null || true)"
if [[ -z "${server}" ]]; then
  echo "skipped: kubectl has no active cluster context"
  exit 0
fi

if [[ "${server}" == "http://localhost:8080" || "${server}" == "https://localhost:8080" ]]; then
  echo "skipped: kubectl points to default localhost:8080 with no active cluster"
  exit 0
fi

kubectl apply --dry-run=client --validate=false -f "${RENDERED_PATH}" >/dev/null
echo "dry-run=client succeeded for ${RENDERED_PATH}"
