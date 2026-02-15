"use client";

import React from "react";
import { TradingViewBar, TradingViewChart, TradingViewHover } from "@/components/market/tradingview-chart";

export type ChartPanelProps = {
  symbol: string;
  quoteCurrency: string;
  bars: TradingViewBar[];
  liveBar: TradingViewBar | null;
  onHover: (hover: TradingViewHover | null) => void;
  onMouseLeave: () => void;
  chartHeaderPriceText: string;
  liveChangeText: string;
  liveChangeToneClass: string;
  openText: string;
  closeText: string;
  initialBarSpacing?: number;
  initialScrollPosition?: number;
  onViewChange?: (view: { barSpacing: number; scrollPosition: number }) => void;
  height?: number | string;
};

export function ChartPanel({
  symbol,
  quoteCurrency,
  bars,
  liveBar,
  onHover,
  onMouseLeave,
  chartHeaderPriceText,
  liveChangeText,
  liveChangeToneClass,
  openText,
  closeText,
  initialBarSpacing,
  initialScrollPosition,
  onViewChange,
  height = "100%"
}: ChartPanelProps) {
  return (
    <div data-testid="chart-panel" className="h-full min-h-0 p-0">
      {bars.length > 0 ? (
        <div
          data-testid="chart-surface"
          className="relative h-full overflow-hidden bg-[#0B0E11]"
          onMouseLeave={onMouseLeave}
        >
          <div className="pointer-events-none absolute left-3 top-3 z-10 flex flex-wrap items-center gap-3 font-mono text-xs text-foreground">
            <span data-testid="chart-header-price">{chartHeaderPriceText}</span>
            <span className="text-muted-foreground">|</span>
            <span data-testid="chart-live-change" className={liveChangeToneClass}>
              {liveChangeText}
            </span>
            <span className="text-muted-foreground">|</span>
            <span data-testid="chart-live-open" className="text-emerald-400">
              {openText}
            </span>
            <span className="text-muted-foreground">|</span>
            <span data-testid="chart-live-close" className="text-rose-400">
              {closeText}
            </span>
          </div>
          <TradingViewChart
            quoteCurrency={quoteCurrency}
            bars={bars}
            liveBar={liveBar}
            onHover={onHover}
            height={height}
            initialBarSpacing={initialBarSpacing}
            initialScrollPosition={initialScrollPosition}
            onViewChange={onViewChange}
          />
        </div>
      ) : (
        <div className="flex h-full items-center justify-center bg-background/40">
          <p className="text-sm text-muted-foreground">No candles yet.</p>
        </div>
      )}
    </div>
  );
}
