# Kalency v1 Rollout and Assumptions

## Implementation Sequence
1. Scaffold repository structure, shared contracts, and environment config.
2. Implement identity/auth with PostgreSQL-backed users and API keys.
3. Implement matching engine and Redis atomic commit path.
4. Implement gateway REST/WS and risk controls.
5. Implement market simulator and candle aggregation pipeline.
6. Implement async ledger writer from Redis Streams to PostgreSQL.
7. Implement Next.js web UI.
8. Add observability stack and runbooks.
9. Tune and validate load gate.

## Deployment Profiles
- `dev`: local development profile.
- `loadtest`: high-load benchmarking profile.
- `prodlike`: production-style validation profile.

## Rollout Path
1. Validate all correctness suites in `dev`.
2. Run load tests in `loadtest` and tune bottlenecks.
3. Execute failure and DR drills in `prodlike`.
4. Launch with controlled synthetic traffic, then scale to target users.

## Functional Defaults
- Trading mode: spot only.
- Order types: market and limit.
- Order actions: place, partial fill, cancel.
- Timeframes: 1s, 5s, 1m, 5m, 1h.
- Data retention: 30 days hot in Redis, older data archived to object storage.

## Data and Consistency Defaults
- Redis is the hot-path source for execution state.
- PostgreSQL stores identity and append-only audit ledger.
- Order acknowledgement occurs after Redis commit, before SQL write completion.
- SQL persistence is async via stream consumers.
- If SQL is down, matching continues with buffered events and active alerting.

## Operational Defaults
- Throughput target: 50,000 orders/sec.
- Latency target: p99 under 20 ms.
- Hard acceptance duration: 15 minutes sustained.
- DR objectives: RPO 5 minutes, RTO 15 minutes.

## Assumptions
1. `docs/kalency-v1` is the current reference set and should stay aligned with active runtime APIs.
2. Existing Redis learning docs remain in `docs/learn-redis/` and are unchanged.
3. Plan version is fixed at `Kalency v1.0 (2026-02-14)` until an explicit revision.
4. New ambiguities must be tracked as `TODO` notes instead of silently changing scope.
