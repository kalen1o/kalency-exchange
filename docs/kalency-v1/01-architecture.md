# Kalency v1 Architecture

## Service Topology
Kalency v1 is a single-region, high-throughput paper exchange with fake buy/sell and real order matching. The primary stack is Go + Fiber + go-redis for hot-path performance.

Core services:
- `gateway-api` (Go/Fiber): REST, WebSocket trades stream, JWT/API-key auth, request routing.
- `matching-engine` (Go): separate deployable service for order matching, wallet/risk checks, and trade execution.
- `market-sim` (Go): synthetic market tick generation.
- `candle-aggregator` (Go): timeframe rollups for 1s, 5s, 1m, 5m, 1h.
- `ledger-writer` (Go): stream consumer that writes append-only trade ledger to PostgreSQL.
- `web` (Next.js): user-facing exchange UI.

## Runtime Data Flow
1. User submits order to `gateway-api`.
2. Gateway authenticates and validates risk limits.
3. Gateway routes command to matching shard by symbol hash.
4. Matching engine updates order/wallet state atomically in Redis and emits execution events to Redis Streams.
5. WebSocket fanout distributes order/trade/book updates.
6. `ledger-writer` asynchronously persists execution events to PostgreSQL.
7. `market-sim` emits ticks and `candle-aggregator` maintains timeframe candles.

## Matching Path
- Spot trading only, no leverage.
- Supported order types: market and limit.
- Lifecycle: place, partial fill, cancel.
- Matching rule: price-time priority.
- Consistency: strict atomic consistency for execution and wallet updates.

## Chart Rendering Path
- Web UI renders charts client-side from candle/trade data served by `gateway-api`.
- Chart interactivity is handled in-browser; no server-side chart render service is required.

## Deployment Shape
- Docker Compose profiles: `dev`, `loadtest`, `prodlike`.
- Current dev compose (`docker/compose.yaml`) runs `redis`, `postgres`, `matching-engine`, `market-sim`, `candle-aggregator`, `ledger-writer`, `gateway-api`, and `web`.
- Redis Cluster baseline: 6 masters + 6 replicas.
- PostgreSQL for identity and audit ledger.
- Object storage for archived data.

For endpoint and message contracts, see `docs/kalency-v1/02-api-and-contracts.md`.
