import { describe, expect, it } from "vitest";
import { Candle } from "./api";
import { renderCandlestickSVG } from "./chart";

const sampleCandles: Candle[] = [
  {
    symbol: "BTC-USD",
    timeframe: "1m",
    bucketStart: "2026-02-15T00:00:00Z",
    open: 100,
    high: 104,
    low: 98,
    close: 102,
    volume: 10
  },
  {
    symbol: "BTC-USD",
    timeframe: "1m",
    bucketStart: "2026-02-15T00:01:00Z",
    open: 102,
    high: 105,
    low: 101,
    close: 101.5,
    volume: 8
  }
];

describe("renderCandlestickSVG", () => {
  it("returns null for empty candle arrays", () => {
    expect(renderCandlestickSVG("BTC-USD", "1m", [])).toBeNull();
  });

  it("renders candle bodies and wicks from real candle inputs", () => {
    const svg = renderCandlestickSVG("BTC-USD", "1m", sampleCandles, 800, 420);

    expect(svg).toContain("BTC-USD");
    expect(svg).toContain("timeframe 1m");
    expect(svg).toContain("candle-body");
    expect(svg).toContain("candle-wick");
    expect(svg).not.toContain("<polyline");
  });
});
