"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type { Candle, OpenOrder, Trade, ChartRangePreset } from "@/lib/api";
import { fetchCandles, fetchOpenOrders, fetchTrades, rangeFromPreset } from "@/lib/api";

export type UseMarketPollingParams = {
  apiBase: string;
  symbol: string;
  backendTimeframe: string;
  rangePreset: ChartRangePreset;
  refreshKey?: string;
  intervalMs?: number;
  tradesLimit?: number;
  onError?: (message: string | null) => void;
};

export function useMarketPolling({
  apiBase,
  symbol,
  backendTimeframe,
  rangePreset,
  refreshKey,
  intervalMs = 10000,
  tradesLimit = 25,
  onError
}: UseMarketPollingParams) {
  const [orders, setOrders] = useState<OpenOrder[]>([]);
  const [trades, setTrades] = useState<Trade[]>([]);
  const [candles, setCandles] = useState<Candle[]>([]);

  const onErrorRef = useRef<UseMarketPollingParams["onError"]>(onError);
  useEffect(() => {
    onErrorRef.current = onError;
  }, [onError]);

  const refresh = useCallback(async () => {
    try {
      const { from, to } = rangeFromPreset(rangePreset);
      const [openOrders, recentTrades, recentCandles] = await Promise.all([
        fetchOpenOrders(apiBase),
        fetchTrades(apiBase, symbol, tradesLimit),
        fetchCandles(apiBase, symbol, backendTimeframe, from, to)
      ]);
      setOrders(openOrders);
      setTrades(recentTrades);
      setCandles(recentCandles);
      onErrorRef.current?.(null);
    } catch (err) {
      onErrorRef.current?.(err instanceof Error ? err.message : "Failed to refresh data");
    }
  }, [apiBase, symbol, backendTimeframe, rangePreset, tradesLimit]);

  useEffect(() => {
    void refresh();
    const timer = window.setInterval(() => {
      void refresh();
    }, intervalMs);

    return () => window.clearInterval(timer);
  }, [refresh, intervalMs, refreshKey]);

  return { orders, trades, candles, refresh };
}
