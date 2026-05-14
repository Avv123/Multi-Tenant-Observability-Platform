# Local Kubernetes Notes

This repository now includes three local Kubernetes validation paths:

- Helm chart source in `deploy/helm/pulselens`
- Dockerized Helm render script in `scripts/render_helm.sh`
- Client-side manifest dry-run in `scripts/k8s_client_dry_run.sh`
- Optional `k3d` smoke in `scripts/k8s_optional_smoke.sh`
- Local image import helper in `scripts/k3d_load_images.sh`
- Local override values in `deploy/helm/pulselens/values-local-k3d.yaml`

Practical workflow:

1. `bash scripts/render_helm.sh`
2. `bash scripts/k8s_client_dry_run.sh /tmp/pulselens-helm-rendered.yaml`
3. `bash scripts/k8s_optional_smoke.sh`

Behavior:

- if `kubectl` is missing, client dry-run exits with a non-failing `skipped` result
- if `kubectl` exists but has no active cluster context, client dry-run exits with a non-failing `skipped` result
- if `k3d` is missing, the smoke script exits with a non-failing `skipped` result
- the Helm deployments now use `/health` and `/ready` probes
- no tool installation is attempted automatically

This path is packaging validation only. Docker Compose remains the authoritative local runtime.

The chart is still app-only. It assumes backing infra already exists.
