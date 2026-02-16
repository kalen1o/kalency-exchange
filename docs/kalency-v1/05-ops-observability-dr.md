# Kalency v1 Operations, Observability, and DR

## Observability Stack
- Prometheus for metrics collection.
- Grafana for dashboards and SLO views.
- OpenTelemetry for tracing across services.

## Required Dashboards
- API latency (`p50`, `p95`, `p99`) by endpoint.
- Matching throughput and queue depth by symbol shard.
- Redis metrics: ops/sec, memory, evictions, keyspace hit/miss, replication state, slow commands.
- Stream metrics: pending entries and consumer lag.
- PostgreSQL metrics: write latency, error rate, connection saturation.
- WebSocket metrics: active connections, publish rate, drop rate.

## Alerting Baseline
- High p99 latency sustained above threshold.
- Redis memory pressure and eviction spikes.
- Stream backlog growth beyond safe lag budget.
- PostgreSQL write failures or extended downtime.
- WebSocket disconnect spikes.

## Operational Runbook Scope
- Incident triage checklist.
- Redis diagnostics (`INFO`, `SLOWLOG`, latency tools).
- Stream consumer recovery and replay steps.
- PostgreSQL outage protocol with backlog drain procedure.

## Disaster Recovery Targets
- RPO: 5 minutes.
- RTO: 15 minutes.

## DR Mechanisms
- Redis persistence enabled (AOF + RDB).
- Periodic backups with restore drills.
- Stream offsets tracked for deterministic recovery.
- Object storage archive for historical replay/import.

Rollout sequencing and environment progression are defined in `docs/kalency-v1/07-rollout-and-assumptions.md`.
