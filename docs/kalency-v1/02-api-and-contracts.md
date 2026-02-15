# Kalency v1 API and Contracts

This file is the canonical source of public API, event channels, and message/type definitions.

## REST Endpoints (`/v1`)

### Trading
- `POST /v1/orders` place market or limit order.
- `DELETE /v1/orders/{orderId}` cancel open order.
- `GET /v1/orders/open` list open orders for authenticated user.

### Wallet and Account
- `GET /v1/wallet` get paper wallet balances and positions.

### Market Data
- `GET /v1/markets/{symbol}/book` get order book snapshot/depth.
- `GET /v1/markets/{symbol}/trades` get recent trade executions.
- `GET /v1/markets/{symbol}/candles?tf=1s|5s|1m|5m|1h&from=&to=` get OHLCV candles.

### Admin Simulation Controls
- `POST /v1/admin/sim/start`
- `POST /v1/admin/sim/stop`
- `POST /v1/admin/sim/volatility-profile`
- `POST /v1/admin/symbols/{symbol}/pause`
- `POST /v1/admin/symbols/{symbol}/resume`

## WebSocket Channels
- `book.{symbol}`
- `trades.{symbol}`
- `candles.{symbol}.{tf}`
- `orders.{userId}`
- `wallet.{userId}`

## Internal Service Interfaces

### gRPC
- `MatchingEngine.MatchOrder`
- `MatchingEngine.CancelOrder`
- `ChartRenderGateway.RenderChart`
- `ChartRenderGateway.Health`

## Message and Type Definitions

### PlaceOrderRequest
- `clientOrderId`: string
- `symbol`: string
- `side`: enum (`BUY`, `SELL`)
- `type`: enum (`MARKET`, `LIMIT`)
- `price`: decimal (required for limit)
- `qty`: decimal
- `timeInForce`: enum (`GTC`, `IOC`)

### OrderAck
- `orderId`: string
- `status`: enum (`ACCEPTED`, `PARTIALLY_FILLED`, `FILLED`, `CANCELED`, `REJECTED`)
- `filledQty`: decimal
- `remainingQty`: decimal
- `avgPrice`: decimal
- `ts`: RFC3339 timestamp

### ExecutionEvent
- `tradeId`: string
- `buyOrderId`: string
- `sellOrderId`: string
- `symbol`: string
- `price`: decimal
- `qty`: decimal
- `ts`: RFC3339 timestamp

### Candle
- `symbol`: string
- `timeframe`: enum (`1s`, `5s`, `1m`, `5m`, `1h`)
- `bucketStart`: RFC3339 timestamp
- `open`: decimal
- `high`: decimal
- `low`: decimal
- `close`: decimal
- `volume`: decimal

## Contract Notes
- Documentation-only split; no API behavior changes introduced.
- Other plan files may reference these contracts but should not duplicate or redefine them.
