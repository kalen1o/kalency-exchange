# Kalency v1 Testing and Acceptance

## Test Strategy
Kalency v1 requires both correctness and sustained-performance validation.

## Correctness Test Scenarios
- Price-time priority matching behavior.
- Market and limit order execution rules.
- Partial fill progression and remaining quantity accounting.
- Cancel flow for open orders.
- Wallet debit/credit atomicity and non-negative balances.
- Risk checks: rate-limit enforcement, max open orders, position caps.

## Contract and Integration Scenarios
- REST request/response validation for all trading and market routes.
- WebSocket channel subscription and event ordering checks.
- Candle aggregation correctness for 1s, 5s, 1m, 5m, 1h.

## Failure-Mode Scenarios
- PostgreSQL outage while matching continues.
- Redis replica failover and client reconnection behavior.
- Stream consumer restart and idempotent replay from checkpoints.
- Service restart consistency for open orders and wallet snapshots.

## Load and Throughput Scenarios
Tooling: `k6`.

Mandatory pass gate:
- sustain `50,000 orders/sec` for `15 minutes`,
- maintain `p99 < 20 ms`,
- zero integrity failures in wallet/order/ledger invariants.

## Documentation-Split Validation
- Every section from the Kalency master plan is represented in exactly one file under `docs/kalency-v1/`.
- README links resolve to existing files.
- SLO values in `04-performance-and-slos.md` are identical to load-test acceptance values.
- DR values are consistent with `05-ops-observability-dr.md` and `07-rollout-and-assumptions.md`.
