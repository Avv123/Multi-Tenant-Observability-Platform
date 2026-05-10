# Local Kubernetes Notes

This repository now includes two local Kubernetes validation paths:

- Helm chart source in `deploy/helm/pulselens`
- Dockerized Helm render script in `scripts/render_helm.sh`

Current machine state:

- `kubectl` exists
- `helm` is not installed
- `k3d` is not installed
- no current Kubernetes context is configured

Practical workflow:

1. `bash scripts/render_helm.sh`
2. `bash scripts/k8s_client_dry_run.sh`
3. if `k3d` is later installed:
   - create a local cluster
   - install infra dependencies separately
   - apply the rendered manifests or install the chart directly

The chart is still app-only. It assumes backing infra already exists.
