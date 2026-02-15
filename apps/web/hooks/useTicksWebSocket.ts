"use client";

import { useEffect, useRef } from "react";
import type { Candle, ChartInterval, Tick } from "@/lib/api";
import { buildTicksWebSocketURL } from "@/lib/api";
import type { TradingViewBar } from "@/components/market/tradingview-chart";

function timeframeToBucketMs(interval: ChartInterval): number {
  switch (interval) {
    case "1m":
      return 60_000;
    case "5m":
      return 5 * 60_000;
    case "1h":
      return 60 * 60_000;
    default:
      return 60_000;
  }
}

export type UseTicksWebSocketParams = {
  apiBase: string;
  symbol: string;
  timeframe: ChartInterval;
  latestCandle: Candle | null;
  setLiveConnected: (value: boolean) => void;
  setLiveBar: React.Dispatch<React.SetStateAction<TradingViewBar | null>>;
};

export function useTicksWebSocket({ apiBase, symbol, timeframe, latestCandle, setLiveConnected, setLiveBar }: UseTicksWebSocketParams) {
  const tickSilenceTimer = useRef<number | null>(null);
  const lastTickAtRef = useRef<number>(0);
  const latestCandleRef = useRef<Candle | null>(latestCandle);

  useEffect(() => {
    latestCandleRef.current = latestCandle;
  }, [latestCandle]);

  useEffect(() => {
    let ws: WebSocket | null = null;
    let reconnectTimer: number | null = null;
    let cleanedUp = false;

    const connect = () => {
      if (cleanedUp) {
        return;
      }

      ws = new WebSocket(buildTicksWebSocketURL(apiBase, symbol));

      ws.onopen = () => {
        if (cleanedUp) {
          return;
        }
        setLiveConnected(true);

        // Mark as "live" only if ticks continue arriving; if we connect but receive nothing,
        // flip the badge back after a short grace period.
        lastTickAtRef.current = Date.now();
        if (tickSilenceTimer.current !== null) {
          window.clearTimeout(tickSilenceTimer.current);
        }
        tickSilenceTimer.current = window.setTimeout(() => {
          const age = Date.now() - lastTickAtRef.current;
          if (age > 2500) {
            setLiveConnected(false);
          }
        }, 2600);
      };

      ws.onerror = () => {
        if (cleanedUp) {
          return;
        }
        setLiveConnected(false);
      };

      ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(String(event.data));
          if (payload?.type !== "tick") return;
          const tick = payload?.data as Tick | undefined;
          if (!tick || tick.symbol?.toUpperCase?.() !== symbol.toUpperCase()) return;

          const tsMs = Date.parse(String(tick.ts));
          const priceValue = Number((tick as any).price);
          const volumeValue = Number((tick as any).volume ?? 1);
          if (!Number.isFinite(tsMs) || !Number.isFinite(priceValue)) return;

          lastTickAtRef.current = Date.now();
          if (tickSilenceTimer.current !== null) {
            window.clearTimeout(tickSilenceTimer.current);
          }
          tickSilenceTimer.current = window.setTimeout(() => {
            const age = Date.now() - lastTickAtRef.current;
            if (age > 2500) {
              setLiveConnected(false);
            }
          }, 2600);

          const bucketMs = timeframeToBucketMs(timeframe);
          const bucketStartMs = Math.floor(tsMs / bucketMs) * bucketMs;
          const bucketTimeSec = Math.floor(bucketStartMs / 1000) as any;

          setLiveBar((prev) => {
            if (!prev || Number(prev.timeSec) !== Number(bucketTimeSec)) {
              const latest = latestCandleRef.current;
              const seed = latest && Math.floor(Date.parse(latest.bucketStart) / 1000) === Number(bucketTimeSec) ? latest : null;
              const baseOpen = seed ? seed.open : prev ? prev.close : priceValue;
              const baseHigh = seed ? seed.high : baseOpen;
              const baseLow = seed ? seed.low : baseOpen;
              const baseVol = seed ? seed.volume : 0;

              const nextOpen = baseOpen;
              const nextClose = priceValue;
              const nextHigh = Math.max(baseHigh, priceValue);
              const nextLow = Math.min(baseLow, priceValue);
              const nextVol = baseVol + (Number.isFinite(volumeValue) && volumeValue > 0 ? volumeValue : 1);

              return { timeSec: bucketTimeSec, open: nextOpen, high: nextHigh, low: nextLow, close: nextClose, volume: nextVol };
            }

            const nextHigh = Math.max(prev.high, priceValue);
            const nextLow = Math.min(prev.low, priceValue);
            const nextVol = prev.volume + (Number.isFinite(volumeValue) && volumeValue > 0 ? volumeValue : 1);
            return { ...prev, high: nextHigh, low: nextLow, close: priceValue, volume: nextVol };
          });
        } catch {
          return;
        }
      };

      ws.onclose = () => {
        if (cleanedUp) {
          return;
        }
        setLiveConnected(false);
        reconnectTimer = window.setTimeout(connect, 1200);
      };
    };

    connect();

    return () => {
      cleanedUp = true;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
      }
      if (tickSilenceTimer.current !== null) {
        window.clearTimeout(tickSilenceTimer.current);
        tickSilenceTimer.current = null;
      }
      ws?.close();
    };
  }, [apiBase, symbol, timeframe, setLiveConnected, setLiveBar]);
}
