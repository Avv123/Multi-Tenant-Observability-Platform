# Tenant Service (Control Plane)

The `pulselens-tenant-service` acts as the **Control Plane** for the PulseLens observability platform. It is responsible for managing authentication, authorization, API keys, tenants, services, and audit logging. 

Unlike the telemetry plane (which handles millions of logs/metrics via Kafka and ClickHouse), this service handles low-throughput, high-consistency operations backed by PostgreSQL.

---

## 🗄️ Database Tables

The service relies on PostgreSQL to maintain relational consistency. Here is what each table does:

*   **`tenants`**: Represents an isolated organization or billing entity. It holds their subscription plan, ingest quotas, and data retention windows.
*   **`users`**: Represents human accounts that can log into the React UI dashboard. It stores hashed passwords, emails, and their role (e.g., `viewer`, `admin`).
*   **`services`**: Represents the applications (e.g., `api-gateway`, `payment-worker`) that belong to a tenant and are sending telemetry data.
*   **`api_keys`**: Cryptographic tokens assigned to a `service`. The `ingest-service` queries these to authenticate incoming telemetry traffic and enforce rate limits.
*   **`audit_logs`**: An append-only table that tracks "who did what" in the control plane (e.g., "User A revoked API Key B").

---

## 🛣️ API Routes

The router is split into three distinct security zones:

### 1. Health & Infrastructure
*   `GET /health`: Returns 200 OK if the Go process is alive (used by Kubernetes/Docker for liveness).
*   `GET /ready`: Checks if Postgres and Redis are reachable. Returns `ready` or `degraded`.

### 2. Public Authentication (`/api/v1/auth`)
*   `POST /login`: Accepts email/password and returns a signed JWT.
*   **`GET /me`**: *(See specific explanation below)*

### 3. Internal Service-to-Service (`/internal/api/v1`)
Protected by a static `INTERNAL_TOKEN` header. Only other backend services (like `ingest-service` or `alerting-service`) are allowed to call these routes.
*   `POST /auth/resolve-api-key`: Extremely critical route. `ingest-service` calls this to quickly convert an `X-API-Key` string into a valid `TenantID` and `ServiceID` to tag incoming telemetry.
*   *(Includes other CRUD operations used by background workers and CLI bootstrap scripts).*

### 4. Admin Dashboard (`/admin/api/v1`)
Protected by JWT and RBAC (Role-Based Access Control). Called primarily by the React UI.
*   `POST /api-keys/:key_id/rotate`: Rotates a compromised API key.
*   `POST /api-keys/:key_id/revoke`: Disables an API key instantly.
*   `GET /tenants/:tenant_id/audit-logs`: Fetches the history of control plane actions.
*   *(Standard CRUD for Users, Services, and API Keys).*

---

## 🧠 What does `auth.GET("/me", tenantcontrollers.AuthenticateJWT(ctx), ctrl.Me)` do?

You asked specifically about this line in `router.go:L43`. 
This route is used by the frontend (React app) immediately after it boots up or when a user refreshes the page. 

1.  **`AuthenticateJWT(ctx)`**: This is a middleware. Before the request even reaches the controller, it intercepts the `Authorization: Bearer <token>` header, decodes the JWT, verifies the cryptographic signature, and extracts the `user_id` and `tenant_id`. It then injects this identity into the Gin Context.
2.  **`ctrl.Me`**: This is the controller handler. It looks at the injected context, says "Ah, this request belongs to User ID 123", fetches their full profile and permissions from the database, and returns it. 

**Why it's needed:** The frontend uses this to figure out "Who is currently logged in?" and "What UI elements should I show them based on their role?"

---

## ⚙️ Technical Architecture Decisions

There are several non-obvious engineering decisions built into this service, particularly in `main.go`:

### Startup Preflight (`appinit.Preflight`)
Instead of starting the HTTP server immediately and failing requests later, the service runs a "Preflight" check. It actively pings PostgreSQL and Redis on startup. If either is down, the service `panic`s and crashes immediately. This "Fail Fast" mechanism prevents Kubernetes/Docker from routing traffic to a broken container.

### Distributed ID Generation (`idgen.Configure`)
We do not use `AUTO_INCREMENT` or `SERIAL` for database Primary Keys. Instead, the service generates 64-bit Snowflake IDs in memory before hitting the database. This allows us to scale out to multiple `tenant-service` replicas without database locking bottlenecks, and ensures that IDs are globally unique across the entire PulseLens ecosystem.

### Graceful Shutdown (`signal.NotifyContext`)
When you deploy a new version, Docker sends a `SIGTERM` signal to stop the old container. Instead of abruptly killing the process (which would drop in-flight HTTP requests from users), this intercepts the signal and allows the HTTP server up to a few seconds to finish processing current requests before gracefully exiting.

### Service Heartbeats (`platformruntime.Start`)
Instead of hardcoding IPs, `tenant-service` registers itself in Redis and pulses a heartbeat every few seconds. The `query-service` and UI read these heartbeats to dynamically construct a real-time topology map of the cluster, allowing operators to instantly see if `tenant-service` is up, down, or degraded.
