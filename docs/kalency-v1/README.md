# Kalency v1 Plan Index

- Version: Kalency v1.0 (2026-02-14)
- Scope: Kalency v1 architecture and contracts reference (runtime-aligned)
- Canonical API contracts: `docs/kalency-v1/02-api-and-contracts.md`

## Overview
This folder contains the split Kalency v1 planning context. It is separated from `docs/learn-redis/` so architecture, delivery, and operational decisions are easy to find and update without mixing with lesson notes.

## Documents
1. [Architecture](./01-architecture.md)
2. [API and Contracts](./02-api-and-contracts.md)
3. [Data Model: Redis and PostgreSQL](./03-data-model-redis-postgres.md)
4. [Performance and SLOs](./04-performance-and-slos.md)
5. [Operations, Observability, and DR](./05-ops-observability-dr.md)
6. [Testing and Acceptance](./06-testing-and-acceptance.md)
7. [Rollout and Assumptions](./07-rollout-and-assumptions.md)

## Usage Rules
- `02-api-and-contracts.md` is the single source of truth for endpoint, event, and message/type definitions.
- Other files reference API semantics but do not redefine contracts.
- If future ambiguity appears, preserve existing wording and add a `TODO` note instead of introducing a new requirement.

## Split Validation Checklist
- Every Kalency master-plan section appears in one split file.
- README links resolve to local files.
- Throughput and latency SLO values match acceptance criteria.
- DR targets are consistent between operations and rollout docs.
