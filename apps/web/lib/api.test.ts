import { afterEach, describe, expect, it, vi } from "vitest";
import {
  BINANCE_CHART_INTERVALS,
  buildOrderPayload,
  buildTradesWebSocketURL,
  buildTicksWebSocketURL,
  cancelOrder,
  ensureSimulatedSymbol,
  fetchCandles,
  searchPairsByAPI,
  fetchTrades,
  mapChartIntervalToBackendTimeframe,
  rangeFromPreset,
  summarizeTrades
} from "./api";

describe("buildOrderPayload", () => {
  it("omits price for market orders", () => {
    const payload = buildOrderPayload({
      clientOrderId: "c-1",
      userId: "u1",
      symbol: "btc-usd",
      side: "BUY",
      type: "MARKET",
      qty: 2,
      price: 100
    });

    expect(payload.symbol).toBe("BTC-USD");
    expect(payload).not.toHaveProperty("price");
  });

  it("includes price for limit orders", () => {
    const payload = buildOrderPayload({
      clientOrderId: "c-2",
      userId: "u1",
      symbol: "BTC-USD",
      side: "SELL",
      type: "LIMIT",
      qty: 2,
      price: 100
    });

    expect(payload.price).toBe(100);
  });
});

describe("summarizeTrades", () => {
  it("returns last trade and cumulative quantity", () => {
    const result = summarizeTrades([
      { tradeId: "t1", symbol: "BTC-USD", price: 100, qty: 3, ts: "2026-02-14T00:00:00Z" },
      { tradeId: "t2", symbol: "BTC-USD", price: 102, qty: 2, ts: "2026-02-14T00:00:01Z" }
    ]);

    expect(result.lastPrice).toBe(102);
    expect(result.totalQty).toBe(5);
  });
});

describe("market data calls", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls candles endpoint with timeframe/from/to query", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            symbol: "BTC-USD",
            timeframe: "1m",
            bucketStart: "2026-02-15T00:01:00Z",
            open: 100,
            high: 101,
            low: 99,
            close: 100.5,
            volume: 10
          }
        ]),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await fetchCandles("http://localhost:8080", "btc-usd", "1m", "2026-02-15T00:00:00Z", "2026-02-15T00:05:00Z");

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/v1/markets/BTC-USD/candles?tf=1m&from=2026-02-15T00%3A00%3A00Z&to=2026-02-15T00%3A05%3A00Z",
      expect.objectContaining({ headers: expect.any(Object) })
    );
  });

  it("sends auth headers when fetching trades", async () => {
    process.env.NEXT_PUBLIC_API_KEY = "demo-key";
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify([]), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await fetchTrades("http://localhost:8080", "BTC-USD", 10);

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/v1/markets/BTC-USD/trades?limit=10",
      expect.objectContaining({
        headers: expect.objectContaining({ "X-API-Key": "demo-key" })
      })
    );
  });

  it("calls cancel order endpoint with DELETE", async () => {
    process.env.NEXT_PUBLIC_API_KEY = "demo-key";
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          orderId: "ord-1",
          status: "CANCELED",
          filledQty: 0,
          remainingQty: 0,
          avgPrice: 0,
          ts: "2026-02-15T00:00:00Z"
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    await cancelOrder("http://localhost:8080", "ord-1");

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/v1/orders/ord-1",
      expect.objectContaining({
        method: "DELETE",
        headers: expect.objectContaining({ "X-API-Key": "demo-key" })
      })
    );
  });

  it("ensures simulator symbol via admin endpoint", async () => {
    process.env.NEXT_PUBLIC_API_KEY = "demo-key";
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ symbol: "SOL-USD", ensured: true }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await ensureSimulatedSymbol("http://localhost:8080", "sol-usd");

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/v1/admin/symbols/SOL-USD/ensure",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ "X-API-Key": "demo-key" })
      })
    );
  });

  it("searches pair options from API and preserves quote assets", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          symbols: [
            { symbol: "BTCUSDT", status: "TRADING", isSpotTradingAllowed: true },
            { symbol: "ETHBTC", status: "TRADING", isSpotTradingAllowed: true },
            { symbol: "XRPUSDT", status: "BREAK" }
          ]
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await searchPairsByAPI("bt", ["BTC-USD", "ETH-USD", "BTC-ETH"]);

    expect(fetchMock).toHaveBeenCalledWith("https://api.binance.com/api/v3/exchangeInfo");
    expect(result).toContain("BTC-USD");
    expect(result).toContain("ETH-BTC");
    expect(result).toContain("BTC-ETH");
    expect(result).not.toContain("XRP-USD");
  });

  it("returns remote pairs when query is empty", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          symbols: [
            { symbol: "SOLUSDT", status: "TRADING", isSpotTradingAllowed: true },
            { symbol: "ADABTC", status: "TRADING", isSpotTradingAllowed: true }
          ]
        }),
        { status: 200 }
      )
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await searchPairsByAPI("", ["BTC-USD"], 10);

    expect(fetchMock).toHaveBeenCalledWith("https://api.binance.com/api/v3/exchangeInfo");
    expect(result).toContain("BTC-USD");
    expect(result).toContain("SOL-USD");
    expect(result).toContain("ADA-BTC");
  });
});

describe("rangeFromPreset", () => {
  it("returns from/to for 15m preset", () => {
    const now = new Date("2026-02-15T12:00:00.000Z");
    const range = rangeFromPreset("15m", now);
    expect(range.to).toBe("2026-02-15T12:00:00.000Z");
    expect(range.from).toBe("2026-02-15T11:45:00.000Z");
  });
});

describe("binance chart interval mapping", () => {
  it("includes only backend-native ui intervals", () => {
    expect(BINANCE_CHART_INTERVALS).toEqual(["1m", "5m", "1h"]);
  });

  it("maps binance intervals to supported backend timeframes", () => {
    expect(mapChartIntervalToBackendTimeframe("1m")).toBe("1m");
    expect(mapChartIntervalToBackendTimeframe("3m")).toBe("1m");
    expect(mapChartIntervalToBackendTimeframe("5m")).toBe("5m");
    expect(mapChartIntervalToBackendTimeframe("15m")).toBe("5m");
    expect(mapChartIntervalToBackendTimeframe("1h")).toBe("1h");
    expect(mapChartIntervalToBackendTimeframe("4h")).toBe("1h");
    expect(mapChartIntervalToBackendTimeframe("1d")).toBe("1h");
    expect(mapChartIntervalToBackendTimeframe("1w")).toBe("1h");
    expect(mapChartIntervalToBackendTimeframe("1M")).toBe("1h");
  });
});

describe("buildTradesWebSocketURL", () => {
  it("builds ws URL from http API base", () => {
    expect(buildTradesWebSocketURL("http://localhost:8080", " btc-usd ")).toBe("ws://localhost:8080/ws/trades/BTC-USD");
  });

  it("builds wss URL from https API base", () => {
    expect(buildTradesWebSocketURL("https://api.example.com/", "eth-usd")).toBe("wss://api.example.com/ws/trades/ETH-USD");
  });
});

describe("buildTicksWebSocketURL", () => {
  it("builds ws URL from http API base", () => {
    expect(buildTicksWebSocketURL("http://localhost:8080", " btc-usd ")).toBe("ws://localhost:8080/ws/ticks/BTC-USD");
  });

  it("builds wss URL from https API base", () => {
    expect(buildTicksWebSocketURL("https://api.example.com/", "eth-usd")).toBe("wss://api.example.com/ws/ticks/ETH-USD");
  });
});
