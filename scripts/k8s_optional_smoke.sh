#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLUSTER_NAME="${1:-pulselens-local}"
RENDERED_PATH="${2:-/tmp/pulselens-helm-rendered.yaml}"
RELEASE_NAME="pulselens"

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

if ! command -v helm >/dev/null 2>&1; then
  echo "skipped: helm not installed locally"
  exit 0
fi

cleanup() {
  helm uninstall "${RELEASE_NAME}" >/dev/null 2>&1 || true
  k3d cluster delete "${CLUSTER_NAME}" >/dev/null 2>&1 || true
}

trap cleanup EXIT

k3d cluster create "${CLUSTER_NAME}" --agents 1 >/dev/null
"${ROOT_DIR}/scripts/k3d_load_images.sh" "${CLUSTER_NAME}"
helm install "${RELEASE_NAME}" "${ROOT_DIR}/deploy/helm/pulselens" -f "${ROOT_DIR}/deploy/helm/pulselens/values-local-k3d.yaml" >/dev/null
kubectl wait --for=condition=available --timeout=180s deployment/pulselens-tenant deployment/pulselens-ingest deployment/pulselens-processing deployment/pulselens-query deployment/pulselens-alerting deployment/pulselens-archive >/dev/null
kubectl get pods >/dev/null
kubectl get svc >/dev/null

echo "k3d smoke succeeded for ${CLUSTER_NAME}"
