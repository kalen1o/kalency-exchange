# Kalency v1 Performance and SLOs

## Throughput and Latency Targets
- Primary throughput target: `50,000 orders/sec`.
- Latency objective: `p99 < 20 ms` on order placement path.
- Active user target: up to `20,000` concurrent active users.
- Symbol scope: up to `200` trading pairs.

## Hard Acceptance Gate
Kalency v1 is accepted only if it sustains:
- `50,000 orders/sec`
- `p99 < 20 ms`
- duration: `15 continuous minutes`
- no data-integrity violations (wallet/order consistency preserved)

## Capacity and Topology Baseline
- Redis Cluster: `6 masters + 6 replicas`.
- Single-region deployment for v1.
- Hot read/write path optimized around Redis.
- SQL writes remain async to protect latency budgets.

## Performance Design Choices
- Symbol-based partitioning for matching shards.
- Sequential per-symbol matching loop to maintain deterministic order.
- Atomic state updates in Redis to avoid read-modify-write races.
- Stream-driven async ledger persistence.
- Redis pipelining for non-critical batched operations.

## Resource Planning Inputs
- Timeframes served: `1s`, `5s`, `1m`, `5m`, `1h`.
- 30-day hot data retention in Redis for high-speed queries.
- Chart render path uses cache-first policy to avoid GPU bottlenecks on repeat requests.

## SLO Reporting
Latency SLOs are tracked per endpoint and per processing stage:
- ingress validation,
- matching,
- state commit,
- event fanout.

Acceptance tests and SLO validation scenarios are defined in `docs/kalency-v1/06-testing-and-acceptance.md`.
