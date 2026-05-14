# System Design Notes

This document maps the implemented system-design concerns to the current codebase.

## Implemented Concepts

### Multi-Tenant Isolation

- every ingest envelope carries `tenant_id` and `service_id`
- JWT claims scope query and alerting reads to a tenant
- API keys resolve to exactly one tenant and one service

### Async Processing

- writes are accepted by ingest, then processed asynchronously through Kafka
- the read path is eventually consistent by design
- processing now mirrors hot telemetry into ClickHouse while keeping PostgreSQL as the compatibility store

### Consistent Hashing

- rendezvous hashing is implemented in `libs/pulselens-platform/sharding`
- ingest assigns a stable `shard_id` onto envelopes
- this creates a future-ready partitioning signal without adding a new routing service today

### Rate Limiting

- Redis fixed-window limiting is implemented in `libs/pulselens-platform/ratelimit`
- ingest applies it per `tenant_id + service_id`
- the limiter now supports batch-sized reservations instead of counting only one request at a time

### Quota Enforcement

- Redis daily quota reservation is implemented in `libs/pulselens-platform/quota`
- ingest reserves daily capacity from the tenant’s configured ingest quota before publish

### Idempotency and Duplicate Handling

- processing uses Redis `SETNX` dedupe on `tenant_id + event_id`
- the pipeline is `at-least-once + dedupe`, not exactly-once

### Failure Handling

- processing retries failed events by republishing to a retry topic
- scheduled retries are converted into persisted retry events with `next_attempt_at`
- retry dispatcher republishes due retry events back to the primary Kafka topics
- after max retry count, events are persisted into `dead_letter_events`
- alerting evaluation failures are logged without crashing the loop
- processing cleanup worker continuously removes expired hot-path data and archive metadata
- query service returns an explicit dependency-unavailable error if ClickHouse is down for telemetry reads
- notification-path failure is explicitly exercised by the failure-drill script and recorded in delivery state

### Race Condition Handling

- Redis atomic increment is used for rate limiting and quota counters
- Redis `SETNX` is used for dedupe
- Redis lock ownership is used in alerting so only one evaluator worker acts per cycle

### Distributed Locking

- implemented in `libs/pulselens-platform/lock`
- alerting uses owner-aware release so one worker does not accidentally release another worker’s lock
- archive replay worker also uses owner-aware Redis locking

### Unique ID Generation

- Snowflake-style sortable IDs are generated in `libs/pulselens-platform/idgen`
- the generator uses a service-derived node ID plus per-millisecond sequence state
- IDs are prefixed and lexically time-ordered for operational readability:
  - `tenant_*`
  - `svc_*`
  - `evt_*`
  - `rule_*`
  - `incident_*`

### Auditability

- tenant control-plane actions create audit rows
- current actions logged:
  - tenant created
  - service created
  - API key created

### Service Health Aggregation

- query service derives service health from recent log severity and recent telemetry timestamps
- current health states:
  - `healthy`
  - `warning`
  - `critical`

### Alerting and Failure Detection

- alert rules evaluate recent telemetry windows
- rules support:
  - `log`
  - `metric`
  - `trace`
- incidents open on breach and resolve on recovery
- incidents can also be acknowledged manually
- incidents can be assigned to tenant operators
- open incidents can escalate automatically if they remain unacknowledged
- incident comments provide lightweight workflow history
- every service emits Redis heartbeats with TTL-backed freshness
- query service projects those heartbeats into `healthy`, `degraded`, and `down` runtime states
- incident notifications now support webhook, Slack-webhook-style, and SMTP email delivery
- per-delivery response payloads are stored for operational debugging
- webhook delivery now retries transient failures before marking a delivery failed

### Archive and Replay

- processing writes archived raw envelopes into S3-compatible object storage
- archive metadata is indexed in PostgreSQL for replay queries and retention cleanup
- replay jobs read archived objects back from object storage and republish with new replay event IDs
- replay worker runs under distributed lock

### Backpressure Control

- ingest reserves pending queue capacity in Redis before Kafka publish
- each queue has an explicit pending threshold
- processing releases pending capacity only on terminal success or DLQ
- query service exposes backlog snapshots so overload is visible from the UI

### Hot Store and Query Caching

- ClickHouse is the hot telemetry read store for logs, metrics, traces, and health aggregations
- processing writes telemetry into PostgreSQL and ClickHouse in the same worker flow
- query service treats ClickHouse as authoritative for telemetry reads and returns degraded-mode errors if ClickHouse is unavailable
- ClickHouse materialized views now maintain minute rollups for telemetry counts, service health, log severity, metrics, and trace latency
- query-side aggregate APIs now read ClickHouse rollup tables first instead of PostgreSQL rollup tables
- query service caches read-heavy responses in Redis with short TTL to reduce repeated scan cost
- cache keys include tenant-scoped cache versions, so ingest, cleanup, saved-query writes, and dashboard writes invalidate stale entries without key scans
- minute rollups in PostgreSQL reduce repeated aggregate scans for overview counts, service-health summaries, log severity, metric series, and trace latency
- logs, metrics, traces, and analytics rollups now each use signal-aware filter handling instead of one generic filter surface

### Infra Packaging

- the repo now owns a complete `docker-compose` stack definition
- every service has a container build definition
- Compose is the only supported local runtime for the hardening path
- port ownership is explicit and enforced through preflight checks
- optional Helm assets exist for local Kubernetes experimentation
- dependency health and Kafka lag are projected into the query/API layer so infra state is observable from the product itself
- Kubernetes validation stays offline-first:
  - Helm render
  - structural manifest sanity checks
  - optional `kubectl` dry-run when a real context exists
  - optional `k3d` smoke only when local tools are already installed

### Retention and Cleanup

- processing owns retention cleanup because it owns the telemetry persistence plane
- cleanup runs now derive per-tenant retention windows from tenant metadata
- cleanup deletes expired telemetry, rollups, retry rows, DLQ rows, and archive metadata
- archive objects are removed only when no archive records still reference that object
- every cleanup execution is persisted into `cleanup_runs` for auditability

### Saved State on Read Side

- saved queries persist common read filters
- dashboards persist widget and layout definitions
- dashboard widgets normalize filters by dataset so preview, saved state, and rendering share the same behavior
- this moves the UI from transient-only state into reusable stored observability views

### Role-Based Access Control

- roles are enforced at route level
- current roles:
  - `tenant_admin`
  - `viewer`
  - `operator`
  - `alert_manager`
  - `service_owner`
- tenant service supports user creation and listing so role enforcement is operational, not only theoretical

### Performance Verification

- the repo includes a concurrent local load-test harness
- the repo now includes a sustained soak-test harness
- the repo now includes a failure-drill harness for broken notification paths
- the latest verified run completed with:
  - `120` ingest requests
  - `40` query requests
  - `0` failures
- a concurrency bug in Redis backpressure reservation was fixed by replacing optimistic transaction-based reservation with atomic counter updates

### Notification Workflow

- notification channels are tenant-scoped
- notification deliveries are recorded for incident lifecycle events
- current local implementation records successful deliveries without depending on paid external providers

## Current Consistency Model

- control-plane writes are strongly consistent at the local database level
- telemetry visibility is eventually consistent because of Kafka and async workers
- alerts are also eventually consistent because they depend on processed telemetry

## Scalability Path

### Works Today

- stateless HTTP services
- Kafka consumer groups
- Redis-backed shared coordination
- independent control-plane and data-plane services
- Redis-backed runtime heartbeats
- Redis-backed pending queue controls

### Next Scale Steps

- increase Kafka partition counts
- move delayed retry orchestration from local scheduler shape to broker-native or queue-native delayed delivery
- add finer-grained cache invalidation beyond the current tenant- and dashboard-scoped versioning
- add horizontal worker groups for separate signal types

## Deliberate Non-Goals In This Local Version

- exactly-once delivery
- globally ordered IDs
- cross-region failover
- full SSO/MFA identity layer
- billing and commercial plan enforcement

## Local Runtime Contract

- PulseLens runs locally through `docker compose`, not a mixed host-process model
- `/health` proves process liveness
- `/ready` proves dependency usability for real traffic
- startup conflicts are reported through `scripts/check_required_ports.sh`
- optional host-run binaries may still exist for development, but they are not part of the supported hardening or validation path

## Why This Is Still Production-Shaped

Even without the expensive production infra, the system already exercises the important architecture concerns:

- tenant isolation
- decoupled ingestion
- async processing
- operational locking
- quotas and rate controls
- retry and DLQ
- rollup-based aggregate reads
- tenant-aware retention
- object-store-backed archive and replay
- webhook notifications
- alert evaluation
- structured package boundaries

## Local Hardening Delta

The local-first hardening layer now adds:

- explicit liveness vs readiness separation
- dependency preflight checks before service boot
- backup and restore drills for control-plane, hot-store, and archive data
- threshold-based local SLO gates for load and soak tests
- Kafka, Postgres, and MinIO chaos drills
- API key rotation and revocation as the local security baseline
