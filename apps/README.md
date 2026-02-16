# Kalency Apps (Initial Implementation)

This directory starts the Kalency application implementation from the docs plan.

## App layout
- Gateway API (Fiber): `apps/gateway-api/cmd/gateway-api`
- Matching engine service: `apps/matching-engine/cmd/matching-engine`
- Market simulator service: `apps/market-sim/cmd/market-sim`
- Candle aggregator service: `apps/candle-aggregator/cmd/candle-aggregator`
- Ledger writer service: `apps/ledger-writer/cmd/ledger-writer`
- Web frontend: `apps/web`

## Implemented so far
- Go module bootstrap.
- In-memory matching engine with:
  - market and limit orders,
  - price-time priority,
  - partial fill support,
  - open-order tracking,
  - execution log,
  - wallet and paper-trading risk checks (quote/base balance constraints).
- Optional Redis-backed open-order read/write path.
- Optional Redis Streams execution-event publishing path.
- Optional Redis Streams trade-read path for market trade queries.
- Market simulator service with:
  - synthetic tick generation for configured symbols,
  - optional Redis Streams tick publishing (`kalency:v1:stream:ticks`),
  - admin controls:
    - `POST /v1/admin/sim/start`
    - `POST /v1/admin/sim/stop`
    - `POST /v1/admin/sim/volatility-profile`,
    - `POST /v1/admin/symbols/{symbol}/pause`
    - `POST /v1/admin/symbols/{symbol}/resume`
  - `GET /healthz`.
- Candle aggregator service with:
  - Redis Streams tick consumption (`kalency:v1:stream:ticks`),
  - Redis candle rollups for `1s`, `5s`, `1m`, `5m`, `1h`,
  - candle key format: `v1:candle:{symbol}:{tf}:{bucketStart}`,
  - 30-day candle TTL default,
  - `GET /healthz`.
- Ledger writer service with:
  - Redis Streams execution consumption (`kalency:v1:stream:executions`),
  - async writes to PostgreSQL `trade_ledger`,
  - idempotent insert by `trade_id`,
  - `GET /healthz`.
- Gateway API endpoints with JWT/API-key auth:
  - `POST /v1/auth/token`
  - `POST /v1/orders`
  - `DELETE /v1/orders/{orderId}`
  - `GET /v1/orders/open`
  - `GET /v1/wallet`
  - `POST /v1/admin/sim/start`
  - `POST /v1/admin/sim/stop`
  - `POST /v1/admin/sim/volatility-profile`
  - `POST /v1/admin/symbols/{symbol}/pause`
  - `POST /v1/admin/symbols/{symbol}/resume`
  - `GET /v1/markets/{symbol}/book`
  - `GET /v1/markets/{symbol}/trades`
  - `GET /v1/markets/{symbol}/candles?tf=1s|5s|1m|5m|1h&from=&to=`
  - `GET /ws/trades/{symbol}`
  - `GET /healthz`
- Next.js web frontend (`apps/web`) with:
  - order form,
  - open orders view,
  - recent trades view,
  - candles view (`/v1/markets/{symbol}/candles`),
  - client-side interactive chart rendering.
- Unit tests for matching and HTTP endpoints.
- Redis store tests using `miniredis`.

## Run tests
```bash
cd apps/gateway-api
go test ./...
```

```bash
cd apps/matching-engine
go test ./...
```

```bash
cd apps/market-sim
go test ./...
```

```bash
cd apps/candle-aggregator
go test ./...
```

```bash
cd apps/ledger-writer
go test ./...
```

```bash
cd apps/web
npm test
```

## Run server
```bash
cd apps/gateway-api
go run ./cmd/gateway-api
```

## Run matching-engine
```bash
cd apps/matching-engine
REDIS_ADDR=127.0.0.1:6379 PORT=8081 go run ./cmd/matching-engine
```

## Run market-sim
```bash
cd apps/market-sim
REDIS_ADDR=127.0.0.1:6379 PORT=8082 SIM_START_ON_BOOT=true SIM_SELL_BIAS=0.65 go run ./cmd/market-sim
```

## Run candle-aggregator
```bash
cd apps/candle-aggregator
REDIS_ADDR=127.0.0.1:6379 PORT=8083 CANDLE_TICK_STREAM=kalency:v1:stream:ticks CANDLE_KEY_PREFIX=v1 go run ./cmd/candle-aggregator
```

## Run ledger-writer
```bash
cd apps/ledger-writer
REDIS_ADDR=127.0.0.1:6379 PORT=8084 LEDGER_STREAM_KEY=kalency:v1:stream:executions POSTGRES_DSN='postgres://kalency:kalency@127.0.0.1:5432/kalency?sslmode=disable' go run ./cmd/ledger-writer
```

## Run gateway-api against matching-engine
```bash
cd apps/gateway-api
MATCHING_ENGINE_URL=http://127.0.0.1:8081 MARKET_SIM_URL=http://127.0.0.1:8082 CANDLE_REDIS_ADDR=127.0.0.1:6379 CANDLE_KEY_PREFIX=v1 JWT_SECRET=dev-secret API_KEYS=demo-key:demo-user PORT=8080 go run ./cmd/gateway-api
```

## Run Docker Compose dev profile
```bash
docker compose -f docker/compose.yaml --profile dev up --build
```

Then open `http://localhost:3000`.

## Run Docker Compose with Makefile
```bash
make up
```

Useful commands:
- `make down`
- `make logs`
- `make ps`
- `make config`
