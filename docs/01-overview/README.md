# Overview

PulseLens is a multi-tenant observability platform built to handle logs, metrics, and traces through a realistic service-oriented backend. The project focuses on strong engineering fundamentals: tenant isolation, asynchronous ingestion, telemetry processing, operational visibility, and recovery workflows.

## What PulseLens Demonstrates

- control-plane and telemetry-plane separation
- event-driven ingestion and processing through Kafka
- ClickHouse-backed telemetry reads and rollups
- tenant-safe querying and dashboard workflows
- alerting, incidents, delivery tracking, and archive/replay
- local production-style runtime hardening with backup, restore, and chaos drills

## Project Constraints

- zero recurring infrastructure cost for the primary runtime
- architecture style kept close to Omniful-style Go services
- end result should behave like a production-shaped backend, not a toy demo

## What To Read Next

- project overview and quick start: `README.md`
- local runtime and testing: `docs/06-local-development/README.md`
- implemented features inventory: `docs/08-implemented-system/README.md`
- architecture rationale: `docs/09-system-design/README.md`
