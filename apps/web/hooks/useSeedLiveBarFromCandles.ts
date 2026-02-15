"use client";

import { useEffect } from "react";
import type { Candle } from "@/lib/api";
import type { TradingViewBar } from "@/components/market/tradingview-chart";

export function useSeedLiveBarFromCandles(sortedCandles: Candle[], setLiveBar: (bar: TradingViewBar | null) => void) {
  useEffect(() => {
    if (sortedCandles.length === 0) {
      setLiveBar(null);
      return;
    }

    const last = sortedCandles[sortedCandles.length - 1];
    const timeSec = Math.floor(Date.parse(last.bucketStart) / 1000);
    if (!Number.isFinite(timeSec)) {
      setLiveBar(null);
      return;
    }

    setLiveBar({
      timeSec: timeSec as any,
      open: last.open,
      high: last.high,
      low: last.low,
      close: last.close,
      volume: last.volume
    });
  }, [sortedCandles, setLiveBar]);
}

