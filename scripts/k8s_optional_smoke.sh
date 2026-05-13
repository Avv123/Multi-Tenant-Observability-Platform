#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLUSTER_NAME="${1:-pulselens-local}"
RENDERED_PATH="${2:-/tmp/pulselens-helm-rendered.yaml}"

if ! command -v k3d >/dev/null 2>&1; then
  echo "skipped: k3d not installed"
  exit 0
fi

if ! command -v kubectl >/dev/null 2>&1; then
  echo "skipped: kubectl not installed"
  exit 0
fi

if [[ ! -f "${RENDERED_PATH}" ]]; then
  "${ROOT_DIR}/scripts/render_helm.sh" "${RENDERED_PATH}"
fi

cleanup() {
  k3d cluster delete "${CLUSTER_NAME}" >/dev/null 2>&1 || true
}

trap cleanup EXIT

k3d cluster create "${CLUSTER_NAME}" --agents 1 >/dev/null
kubectl apply --dry-run=client --validate=false -f "${RENDERED_PATH}" >/dev/null
kubectl apply -f "${RENDERED_PATH}" >/dev/null
kubectl get namespace >/dev/null

echo "k3d smoke succeeded for ${CLUSTER_NAME}"
