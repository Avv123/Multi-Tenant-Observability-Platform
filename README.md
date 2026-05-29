# PulseLens 🛡️
**Enterprise-Grade, Multi-Tenant Observability Platform**

PulseLens is a comprehensive observability suite designed for high-performance telemetry processing and seamless multi-tenant operations. It provides a unified experience for **Logs, Metrics, and Traces**, with built-in support for real-time alerting.

---

### 🚀 Key Features
- **Multi-Tenancy:** Strict data isolation between tenants via the Control Plane.
- **Unified Telemetry:** Native support for structured Logs, numeric Metrics, and distributed Traces.
- **Incident Management:** Full lifecycle tracking with assignments, comments, and automated timelines.
- **Intelligent Alerting:** Flexible rule engine with support for Slack/Webhook/Email notifications.
- **Platform Resilience:** Includes automated chaos drills and backup/restore verification scripts.
- **Modern UI:** Premium React-based dashboarding with a sleek, dark-mode-first aesthetic.

---

### 🏗️ Architecture
PulseLens is built as a set of optimized microservices:
- **`tenant-service`**: Identity, RBAC, and API key management (PostgreSQL).
- **`ingest-service`**: High-throughput telemetry ingestion gate (Kafka).
- **`processing-service`**: Real-time stream processing and deduplication.
- **`query-service`**: Analytics engine for ClickHouse-backed telemetry reads.
- **`alerting-service`**: Threshold-based rule evaluation and incident tracking.
- **`ui`**: Unified investigation and management portal.

---

### 🛠️ Tech Stack
- **Backend:** Go (Golang)
- **Frontend:** React, Vite, Vanilla CSS
- **Databases:** PostgreSQL (Control Plane), ClickHouse (Telemetry), Redis (Caching/Locks)
- **Messaging:** Redpanda (Kafka-compatible)

---

### 🚦 Getting Started
1. **Check Prerequisites:** 
   ```bash
   bash scripts/check_required_ports.sh
   ```
2. **Start the Stack:** 
   ```bash
   docker compose up -d --build
   ```
3. **Wait for Health:** 
   ```bash
   bash scripts/wait_for_compose_stack.sh
   ```
4. **Access UI:** Open `http://localhost:3000`

---

### 🧪 Validation & Testing
The repository includes production-grade validation flows to ensure platform reliability:
- `bash scripts/run_full_validation.sh`: Runs the entire automated validation suite.
- **Smoke Tests:** Basic connectivity and data flow verification across all services.
- **Load/Soak Tests:** Performance validation under sustained pressure (Threshold gates).
- **Chaos Drills:** Simulates failures in Kafka, Redis, and ClickHouse to verify automated recovery.
- **Backup & Restore:** Automated validation of data persistence and recovery procedures.

---

### 📂 Documentation
- [Overview](docs/01-overview/README.md)
- [System Architecture](docs/02-system-architecture/README.md)
