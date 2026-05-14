# Security Operations

## API Key Operations
- `GET /admin/api/v1/api-keys`
- `POST /admin/api/v1/api-keys/:key_id/rotate`
- `POST /admin/api/v1/api-keys/:key_id/revoke`

## Behavior
- rotated key returns a fresh secret and deactivates the old key
- revoked key becomes unusable immediately
- ingest rejects inactive keys
- audit logs record key lifecycle actions

## Env Overrides
Config paths can be overridden with `PULSELENS_*` env vars.

Examples:
- `PULSELENS_JWT_SECRET`
- `PULSELENS_INTERNAL_TOKEN`
- `PULSELENS_ARCHIVE_ACCESS_KEY`
- `PULSELENS_ARCHIVE_SECRET_KEY`
