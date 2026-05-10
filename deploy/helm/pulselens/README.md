# Helm Notes

This chart is intentionally app-only.

What it includes:
- config maps for all PulseLens services
- single-replica deployments for the six app services
- cluster services for internal routing

What it assumes exists separately:
- `postgres`
- `redis`
- `kafka`
- `clickhouse`
- `minio`
- `mailhog`

For local Kubernetes work, the intended path is:
- load built images into `k3d`
- install infra dependencies separately
- run `helm install pulselens ./deploy/helm/pulselens`
