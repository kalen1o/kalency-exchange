# Kalency v1 Data Model: Redis and PostgreSQL

## Data Roles
- Redis Cluster is the hot-path execution store for matching, wallet state, live reads, and stream transport.
- PostgreSQL stores identity metadata and append-only audit ledger records.
- Object storage retains archives older than the Redis hot window and chart assets.

## Redis Keyspace Convention
All keys are prefixed with `v1:`.

### Core State
- `v1:wallet:{userId}` (hash)
- `v1:position:{userId}:{symbol}` (hash)
- `v1:order:{orderId}` (hash)
- `v1:orders:open:{userId}` (sorted set)
- `v1:book:snapshot:{symbol}` (string/json)
- `v1:last_price:{symbol}` (string/decimal)

### Candles and History
- `v1:candle:{symbol}:{tf}:{bucketStart}` (hash with TTL)
- Hot retention: 30 days in Redis.
- Archive target: object storage for data older than 30 days.

### Streams and Messaging
- `v1:stream:executions`
- `v1:stream:ticks`
- `v1:stream:ledger`

### Control and Rate Limits
- `v1:ratelimit:{apiKey}:{window}` counters
- `v1:chart:meta:{hash}` chart metadata pointers

## Redis Consistency Model
- Matching engine performs atomic wallet/order updates and event emission.
- Order acknowledgements are returned after Redis atomic commit.
- PostgreSQL ledger persistence is asynchronous via stream consumer.

## PostgreSQL Schema (v1)

### `users`
Stores user identity attributes.

### `api_keys`
Stores API key metadata and hashed secrets for programmatic clients.

### `trade_ledger`
Append-only execution audit table.
Required fields include:
- `trade_id`
- `symbol`
- `buy_user_id`
- `sell_user_id`
- `price`
- `qty`
- `executed_at`

### `ledger_consumer_offsets`
Stores stream consumer checkpoints for replay/recovery.

## Indexing Strategy
- `trade_ledger(symbol, executed_at)`
- `trade_ledger(buy_user_id, executed_at)` and `trade_ledger(sell_user_id, executed_at)`
- `api_keys(key_hash)`

## Outage Behavior
If PostgreSQL is unavailable:
- continue matching and Redis commits,
- buffer ledger events in Redis Streams,
- raise alerts and apply backpressure thresholds.
