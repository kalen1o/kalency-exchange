# Kalency v1 Gap Closure Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Close the concrete contract and rollout gaps so current code aligns with `docs/kalency-v1` API and deployment expectations.

**Architecture:** Extend existing HTTP-based service boundaries without introducing new protocols. Add missing market-data/admin routes end-to-end via minimal client/server interfaces, then expand compose profiles to support `dev`, `loadtest`, and `prodlike` bootstrapping.

**Tech Stack:** Go (Fiber + net/http), Docker Compose, existing Redis/Postgres integrations.

---

### Task 1: Order Book Contract Path (`matching-engine` -> `gateway-api`)

**Files:**
- Modify: `apps/matching-engine/internal/matching/engine.go`
- Modify: `apps/matching-engine/internal/httpapi/server.go`
- Test: `apps/matching-engine/internal/matching/engine_orderbook_test.go`
- Test: `apps/matching-engine/internal/httpapi/server_test.go`
- Modify: `apps/gateway-api/internal/contracts/types.go`
- Modify: `apps/gateway-api/internal/matchingclient/http_client.go`
- Modify: `apps/gateway-api/internal/gatewayapi/server.go`
- Test: `apps/gateway-api/internal/gatewayapi/server_test.go`

**Step 1: Write failing tests**
- Add tests for engine snapshot aggregation/sorting/depth and `GET /v1/markets/{symbol}/book` in matching-engine and gateway.

**Step 2: Run tests to verify RED**
- Run: `cd apps/matching-engine && go test ./internal/matching ./internal/httpapi`
- Run: `cd apps/gateway-api && go test ./internal/gatewayapi`
- Expected: FAIL due to missing order-book types/methods/routes.

**Step 3: Write minimal implementation**
- Add order-book snapshot types and method in engine.
- Add matching-engine HTTP handler for `/v1/markets/{symbol}/book`.
- Add gateway contract types/client method/route to expose the endpoint.

**Step 4: Run tests to verify GREEN**
- Re-run commands from Step 2 and confirm pass.

### Task 2: Admin Symbol Pause/Resume and Gateway Admin Surface

**Files:**
- Modify: `apps/market-sim/internal/sim/generator.go`
- Modify: `apps/market-sim/internal/sim/publisher.go`
- Test: `apps/market-sim/internal/sim/generator_test.go`
- Modify: `apps/market-sim/internal/httpapi/server.go`
- Test: `apps/market-sim/internal/httpapi/server_test.go`
- Add: `apps/gateway-api/internal/marketsimclient/http_client.go`
- Add: `apps/gateway-api/internal/marketsimclient/http_client_test.go`
- Modify: `apps/gateway-api/internal/gatewayapi/server.go`
- Test: `apps/gateway-api/internal/gatewayapi/server_test.go`
- Modify: `apps/gateway-api/cmd/gateway-api/main.go`

**Step 1: Write failing tests**
- Add tests for pause/resume semantics in generator/http API.
- Add tests for gateway admin endpoints and market-sim HTTP client forwarding.

**Step 2: Run tests to verify RED**
- Run: `cd apps/market-sim && go test ./internal/sim ./internal/httpapi`
- Run: `cd apps/gateway-api && go test ./internal/marketsimclient ./internal/gatewayapi`
- Expected: FAIL due to missing methods/routes/package.

**Step 3: Write minimal implementation**
- Add symbol pause/resume controls in simulator.
- Add market-sim endpoints for `/v1/admin/symbols/{symbol}/pause|resume`.
- Expose admin endpoints in gateway and wire `MARKET_SIM_URL` client.

**Step 4: Run tests to verify GREEN**
- Re-run commands from Step 2 and confirm pass.

### Task 3: Deployment Profile Parity

**Files:**
- Modify: `docker/compose.yaml`
- Modify: `Makefile`
- Modify: `apps/README.md`

**Step 1: Write failing validation checks**
- Add shell checks to assert compose contains `dev`, `loadtest`, and `prodlike` profile usage and gateway market-sim wiring.

**Step 2: Run checks to verify RED**
- Run: `rg -n "loadtest|prodlike|MARKET_SIM_URL" docker/compose.yaml Makefile`
- Expected: Missing profile and market-sim env wiring entries.

**Step 3: Write minimal implementation**
- Add compose profiles for all services and gateway dependency/env for market-sim.
- Update README endpoint lists for new routes.

**Step 4: Run checks to verify GREEN**
- Re-run `rg` checks and validate `docker compose -f docker/compose.yaml --profile loadtest config` exits successfully.

### Task 4: Final Verification Gate

**Files:**
- None (verification only)

**Step 1: Run focused service tests**
- `cd apps/matching-engine && go test ./...`
- `cd apps/market-sim && go test ./...`
- `cd apps/gateway-api && go test ./...`

**Step 2: Run compose config verification**
- `docker compose -f docker/compose.yaml --profile dev config >/dev/null`
- `docker compose -f docker/compose.yaml --profile loadtest config >/dev/null`
- `docker compose -f docker/compose.yaml --profile prodlike config >/dev/null`

**Step 3: Report completion with evidence**
- Summarize implemented changes and exact verification commands/results.
