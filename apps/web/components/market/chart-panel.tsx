"use client";

import React from "react";
import { ChartBarStyle, TradingViewBar, TradingViewChart, TradingViewHover } from "@/components/market/tradingview-chart";

export type ChartPanelProps = {
  symbol: string;
  quoteCurrency: string;
  bars: TradingViewBar[];
  liveBar: TradingViewBar | null;
  onHover: (hover: TradingViewHover | null) => void;
  onMouseLeave: () => void;
  chartHeaderPriceText: string;
  changePercentText: string;
  liveChangeToneClass: string;
  openText: string;
  closeText: string;
  highText: string;
  lowText: string;
  volumeText: string;
  showVolume: boolean;
  barStyle: ChartBarStyle;
  initialBarSpacing?: number;
  initialScrollPosition?: number;
  initialPriceFrom?: number;
  initialPriceTo?: number;
  onViewChange?: (view: { barSpacing: number; scrollPosition: number; priceFrom: number | null; priceTo: number | null }) => void;
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
  changePercentText,
  liveChangeToneClass,
  openText,
  closeText,
  highText,
  lowText,
  volumeText,
  showVolume,
  barStyle,
  initialBarSpacing,
  initialScrollPosition,
  initialPriceFrom,
  initialPriceTo,
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
          <div className="pointer-events-none absolute left-3 right-3 top-3 z-10 flex flex-wrap items-center gap-3 font-mono text-xs text-foreground">
            <span data-testid="chart-header-price">{chartHeaderPriceText}</span>
            <span className="text-muted-foreground">|</span>
            <span data-testid="chart-live-open" className={liveChangeToneClass}>
              {openText}
            </span>
            <span className="text-muted-foreground">|</span>
            <span data-testid="chart-live-close" className={liveChangeToneClass}>
              {closeText}
            </span>
            <span className="text-muted-foreground">|</span>
            <span data-testid="chart-live-high" className={liveChangeToneClass}>
              {highText}
            </span>
            <span className="text-muted-foreground">|</span>
            <span data-testid="chart-live-low" className={liveChangeToneClass}>
              {lowText}
            </span>
            {showVolume && (
              <>
                <span className="text-muted-foreground">|</span>
                <span data-testid="chart-live-volume" className={liveChangeToneClass}>
                  {volumeText}
                </span>
              </>
            )}
            <span className="text-muted-foreground">|</span>
            <span data-testid="chart-live-change-pct" className={liveChangeToneClass}>
              {changePercentText}
            </span>
          </div>
          <TradingViewChart
            quoteCurrency={quoteCurrency}
            bars={bars}
            liveBar={liveBar}
            showVolume={showVolume}
            barStyle={barStyle}
            onHover={onHover}
            height={height}
            initialBarSpacing={initialBarSpacing}
            initialScrollPosition={initialScrollPosition}
            initialPriceFrom={initialPriceFrom}
            initialPriceTo={initialPriceTo}
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
