"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
import React from "react";
import {
  BINANCE_CHART_INTERVALS,
  cancelOrder,
  ChartInterval,
  ChartRangePreset,
  mapChartIntervalToBackendTimeframe,
  OrderType,
  placeOrder,
  Side,
  summarizeTrades
} from "@/lib/api";
import { formatLocaleTime } from "@/lib/datetime";
import { formatFloatPrice } from "@/lib/price";
import { TradingViewBar, TradingViewHover } from "@/components/market/tradingview-chart";
import { ChartPanel } from "@/components/market/chart-panel";
import { ExchangeHeader, type HeaderPanelTab } from "@/components/market/exchange-header";
import { PairsSidebar } from "@/components/market/pairs-sidebar";
import { UserPanelDialog, type UserPanelTab } from "@/components/market/user-panel-dialog";
import { useMarketDocumentTitle } from "@/hooks/useMarketDocumentTitle";
import { useMarketPolling } from "@/hooks/useMarketPolling";
import { useResetOnChange } from "@/hooks/useResetOnChange";
import { useSeedLiveBarFromCandles } from "@/hooks/useSeedLiveBarFromCandles";
import { useTicksWebSocket } from "@/hooks/useTicksWebSocket";
import { useToast } from "@/hooks/use-toast";
import { useMarketUrlState } from "@/hooks/useMarketUrlState";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080";
const DEFAULT_USER = process.env.NEXT_PUBLIC_DEFAULT_USER ?? "demo-user";
const RANGE_PRESETS: ChartRangePreset[] = ["15m", "1h", "4h", "24h"];
const PAIR_OPTIONS = ["BTC-USD", "ETH-USD"] as const;

export default function HomeClient() {
  const { toast } = useToast();
  const [userId, setUserId] = useState(DEFAULT_USER);
  const market = useMarketUrlState({
    pairOptions: PAIR_OPTIONS,
    timeframeOptions: BINANCE_CHART_INTERVALS as ChartInterval[],
    defaultPair: "BTC-USD",
    defaultTimeframe: "1m"
  });
  const symbol = market.pair;
  const timeframe = market.timeframe;
  const [rangePreset, setRangePreset] = useState<ChartRangePreset>("1h");
  const [side, setSide] = useState<Side>("BUY");
  const [type, setType] = useState<OrderType>("MARKET");
  const [qty, setQty] = useState(1);
  const [price, setPrice] = useState(100);
  const [busy, setBusy] = useState(false);
  const [cancelOrderID, setCancelOrderID] = useState<string | null>(null);
  const [liveConnected, setLiveConnected] = useState(false);

  const [hoveredBar, setHoveredBar] = useState<TradingViewHover | null>(null);
  const [liveBar, setLiveBar] = useState<TradingViewBar | null>(null);
  const [panelOpen, setPanelOpen] = useState(false);
  const [panelTab, setPanelTab] = useState<UserPanelTab>("order");

  const backendTimeframe = useMemo(() => mapChartIntervalToBackendTimeframe(timeframe), [timeframe]);
  const { orders, trades, candles, refresh } = useMarketPolling({
    apiBase: API_BASE,
    symbol,
    backendTimeframe,
    rangePreset,
    refreshKey: userId,
    intervalMs: 10000,
    tradesLimit: 25,
    onError: (message) => {
      if (!message) return;
      toast({ variant: "destructive", title: "Request Error", description: message });
    }
  });

  const tradeSummary = useMemo(() => summarizeTrades(trades), [trades]);
  const sortedCandles = useMemo(
    () => [...candles].sort((a, b) => Date.parse(a.bucketStart) - Date.parse(b.bucketStart)),
    [candles]
  );
  const chartBars = useMemo(
    () =>
      sortedCandles.map((candle) => ({
        timeSec: Math.floor(Date.parse(candle.bucketStart) / 1000) as any,
        open: candle.open,
        high: candle.high,
        low: candle.low,
        close: candle.close,
        volume: candle.volume
      })),
    [sortedCandles]
  );
  const quoteCurrency = useMemo(() => {
    const parts = symbol.toUpperCase().split("-");
    return parts[1] && parts[1].trim() !== "" ? parts[1].trim() : "USD";
  }, [symbol]);

  function formatPrice(value: number | null | undefined): string {
    return formatFloatPrice(value, quoteCurrency);
  }

  const candleSummary = useMemo(() => {
    if (sortedCandles.length === 0) {
      return { lastClose: null as number | null, totalVolume: 0 };
    }
    return {
      lastClose: sortedCandles[sortedCandles.length - 1].close,
      totalVolume: sortedCandles.reduce((acc, candle) => acc + candle.volume, 0)
    };
  }, [sortedCandles]);

  const latestCandle = sortedCandles.length > 0 ? sortedCandles[sortedCandles.length - 1] : null;
  const activeCandle = hoveredBar
    ? {
        bucketStart: new Date(hoveredBar.timeMs).toISOString(),
        open: hoveredBar.open,
        high: hoveredBar.high,
        low: hoveredBar.low,
        close: hoveredBar.close,
        volume: hoveredBar.volume ?? 0
      }
    : liveBar
      ? {
          bucketStart: new Date(Number(liveBar.timeSec) * 1000).toISOString(),
          open: liveBar.open,
          high: liveBar.high,
          low: liveBar.low,
          close: liveBar.close,
          volume: liveBar.volume
        }
      : latestCandle;
  const liveChartPrice = tradeSummary.lastPrice ?? candleSummary.lastClose;
  const liveCandleChange = activeCandle ? activeCandle.close - activeCandle.open : null;
  const liveCandleChangePct =
    activeCandle && activeCandle.open !== 0 ? ((activeCandle.close - activeCandle.open) / activeCandle.open) * 100 : null;
  const headerLivePrice = tradeSummary.lastPrice ?? liveBar?.close ?? liveChartPrice;
  const chartHeaderPriceText = hoveredBar
    ? `Hover: ${formatPrice(hoveredBar.close)} @ ${formatLocaleTime(hoveredBar.timeMs)}`
    : `Live: ${formatPrice(headerLivePrice)}`;
  const titleChangePct = useMemo(() => {
    const open = liveBar?.open ?? latestCandle?.open ?? null;
    const close = liveBar?.close ?? latestCandle?.close ?? null;
    if (open === null || close === null || !Number.isFinite(open) || !Number.isFinite(close) || open === 0) {
      return null;
    }
    return ((close - open) / open) * 100;
  }, [liveBar, latestCandle]);
  const onChartHover = useCallback((hover: TradingViewHover | null) => {
    setHoveredBar(hover);
  }, []);
  const openUserPanel = useCallback((tab: HeaderPanelTab) => {
    setPanelTab(tab);
    setPanelOpen(true);
  }, []);

  useMarketDocumentTitle({
    symbol,
    price: headerLivePrice,
    changePct: titleChangePct,
    suffix: "Kalency"
  });
  const liveChangeToneClass =
    liveCandleChange === null || Number.isNaN(liveCandleChange)
      ? "text-muted-foreground"
      : liveCandleChange >= 0
        ? "text-emerald-400"
        : "text-rose-400";
  const liveChangeText =
    liveCandleChange === null || Number.isNaN(liveCandleChange)
      ? "Change: -"
      : `Change: ${formatSignedPrice(liveCandleChange)} (${formatSignedPercent(liveCandleChangePct)})`;

  function formatSignedPrice(value: number | null | undefined): string {
    if (value === null || value === undefined || Number.isNaN(value)) {
      return "-";
    }
    const sign = value > 0 ? "+" : value < 0 ? "-" : "";
    return `${sign}${formatPrice(Math.abs(value))}`;
  }

  function formatSignedPercent(value: number | null | undefined): string {
    if (value === null || value === undefined || Number.isNaN(value)) {
      return "-";
    }
    const sign = value > 0 ? "+" : value < 0 ? "-" : "";
    return `${sign}${Math.abs(value).toFixed(2)}%`;
  }

  function clearChartSurfaceHover() {
    setHoveredBar(null);
  }

  useSeedLiveBarFromCandles(sortedCandles, setLiveBar);
  useTicksWebSocket({ apiBase: API_BASE, symbol, timeframe, latestCandle, setLiveConnected, setLiveBar });
  useResetOnChange(clearChartSurfaceHover, [symbol, timeframe, rangePreset]);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);

    try {
      const ack = await placeOrder(API_BASE, {
        clientOrderId: `web-${Date.now()}`,
        userId,
        symbol,
        side,
        type,
        qty,
        price
      });

      toast({ title: "Order Update", description: `Order ${ack.orderId} ${ack.status}` });
      await refresh();
    } catch (err) {
      toast({ variant: "destructive", title: "Order Failed", description: err instanceof Error ? err.message : "Order failed" });
    } finally {
      setBusy(false);
    }
  }

  async function onCancelOrder(orderId: string) {
    setCancelOrderID(orderId);

    try {
      const ack = await cancelOrder(API_BASE, orderId);
      toast({ title: "Order Update", description: `Order ${ack.orderId} ${ack.status}` });
      await refresh();
    } catch (err) {
      toast({ variant: "destructive", title: "Cancel Failed", description: err instanceof Error ? err.message : "Cancel failed" });
    } finally {
      setCancelOrderID(null);
    }
  }

  return (
    <main className="mx-auto flex min-h-screen max-w-[1900px] flex-col bg-[#0B0E11]">
      <ExchangeHeader
        symbol={symbol}
        timeframeOptions={BINANCE_CHART_INTERVALS}
        timeframe={timeframe}
        onTimeframeChange={(value) => market.setTimeframe(value as ChartInterval)}
        userId={userId}
        onOpenPanel={openUserPanel}
      />

      <section
        data-testid="workspace-grid"
        className="grid flex-1 min-h-0 gap-0 lg:grid-cols-[minmax(0,1fr)_260px] lg:divide-x lg:divide-border/70"
      >
        <ChartPanel
          symbol={symbol}
          quoteCurrency={quoteCurrency}
          bars={chartBars}
          liveBar={liveBar}
          onHover={onChartHover}
          onMouseLeave={clearChartSurfaceHover}
          chartHeaderPriceText={chartHeaderPriceText}
          liveChangeText={liveChangeText}
          liveChangeToneClass={liveChangeToneClass}
          openText={`Open: ${formatPrice(activeCandle?.open)}`}
          closeText={`Close: ${formatPrice(activeCandle?.close)}`}
          initialBarSpacing={market.initialView.barSpacing}
          initialScrollPosition={market.initialView.scrollPosition}
          onViewChange={(view) => {
            const round = (value: number) => Math.round(value * 100) / 100;
            const nextBarSpacing = round(view.barSpacing);
            const nextScrollPosition = round(view.scrollPosition);
            if (
              Math.abs(market.view.barSpacing - nextBarSpacing) < 0.01 &&
              Math.abs(market.view.scrollPosition - nextScrollPosition) < 0.01
            ) {
              return;
            }
            market.setView({ barSpacing: nextBarSpacing, scrollPosition: nextScrollPosition });
          }}
        />
        <PairsSidebar pairs={PAIR_OPTIONS} selected={symbol} onSelect={market.setPair} />
      </section>

      <UserPanelDialog
        open={panelOpen}
        onOpenChange={setPanelOpen}
        tab={panelTab}
        onTabChange={setPanelTab}
        symbol={symbol}
        userId={userId}
        onUserIdChange={setUserId}
        quoteCurrency={quoteCurrency}
        rangePresets={RANGE_PRESETS}
        rangePreset={rangePreset}
        onRangePresetChange={(value) => setRangePreset(value as ChartRangePreset)}
        side={side}
        type={type}
        qty={qty}
        price={price}
        onSideChange={setSide}
        onTypeChange={setType}
        onQtyChange={setQty}
        onPriceChange={setPrice}
        busy={busy}
        onSubmit={onSubmit}
        trades={trades.map((t) => ({ tradeId: t.tradeId, price: t.price, qty: t.qty }))}
        tradeSummary={{ lastPrice: formatPrice(tradeSummary.lastPrice), totalQty: tradeSummary.totalQty }}
        orders={orders.map((o) => ({ orderId: o.orderId, side: o.side, symbol: o.symbol, remainingQty: o.remainingQty, qty: o.qty }))}
        cancelOrderID={cancelOrderID}
        onCancelOrder={(id) => void onCancelOrder(id)}
      />
    </main>
  );
}
