"use client";

import React, { useEffect, useMemo, useRef } from "react";
import {
  CandlestickSeries,
  HistogramSeries,
  createChart,
  isUTCTimestamp,
  TickMarkType,
  type IChartApi,
  type IRange,
  type ISeriesApi,
  type Time,
  type UTCTimestamp,
  type CandlestickData,
  type HistogramData,
  type MouseEventParams,
  type SeriesDataItemTypeMap
} from "lightweight-charts";

type TimeSec = UTCTimestamp;

export type TradingViewBar = {
  timeSec: TimeSec;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
};

export type TradingViewHover = {
  timeMs: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number | null;
};

export type TradingViewChartProps = {
  height?: number | string;
  quoteCurrency: string;
  bars: TradingViewBar[];
  liveBar?: TradingViewBar | null;
  onHover?: (hover: TradingViewHover | null) => void;
  initialBarSpacing?: number;
  initialScrollPosition?: number;
  onViewChange?: (view: { barSpacing: number; scrollPosition: number }) => void;
};

export function TradingViewChart({
  height = "100%",
  quoteCurrency: _quoteCurrency,
  bars,
  liveBar,
  onHover,
  initialBarSpacing,
  initialScrollPosition,
  onViewChange
}: TradingViewChartProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const timeTooltipRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candleSeriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const volumeSeriesRef = useRef<ISeriesApi<"Histogram"> | null>(null);
  const onHoverRef = useRef<TradingViewChartProps["onHover"]>(onHover);
  const onViewChangeRef = useRef<TradingViewChartProps["onViewChange"]>(onViewChange);
  const initialViewRef = useRef<{ barSpacing?: number; scrollPosition?: number }>({
    barSpacing: initialBarSpacing,
    scrollPosition: initialScrollPosition
  });
  const visibleLogicalRangeRef = useRef<IRange<number> | null>(null);
  const viewInitializedRef = useRef(false);
  const viewEmitTimerRef = useRef<number | null>(null);
  const pendingViewRef = useRef<{ barSpacing: number; scrollPosition: number } | null>(null);

  const upColor = "#0ECB81";
  const downColor = "#F6465D";

  const locale = useMemo(() => {
    if (typeof navigator === "undefined") return "en-US";
    return (navigator.languages?.[0] || navigator.language || "en-US").trim() || "en-US";
  }, []);

  const hoverTimeFormatter = useMemo(() => {
    return new Intl.DateTimeFormat(locale, {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false
    });
  }, [locale]);

  useEffect(() => {
    onHoverRef.current = onHover;
  }, [onHover]);

  useEffect(() => {
    onViewChangeRef.current = onViewChange;
  }, [onViewChange]);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    const initialView = initialViewRef.current;

	    const chart = createChart(el, {
	      autoSize: true,
	      layout: {
	        background: { color: "#0B0E11" },
	        textColor: "#C9D1D9",
	        fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace"
	      },
	      rightPriceScale: {
	        borderColor: "rgba(255,255,255,0.10)",
	        ticksVisible: true,
	        entireTextOnly: false,
	        scaleMargins: { top: 0.08, bottom: 0.28 }
	      },
	      timeScale: {
	        borderColor: "rgba(255,255,255,0.10)",
	        rightOffset: 6,
	        barSpacing: typeof initialView.barSpacing === "number" && Number.isFinite(initialView.barSpacing) ? initialView.barSpacing : 10,
	        minBarSpacing: 4,
	        timeVisible: true,
	        secondsVisible: false,
	        ticksVisible: true,
	        tickMarkFormatter: (time: Time, tickMarkType: TickMarkType, _locale: string) => {
	          // Only show hour labels on the X axis.
	          // We intentionally ignore minutes to keep the axis clean at small timeframes.
	          void tickMarkType;
	          if (!isUTCTimestamp(time)) return "";
	          const date = new Date(Number(time) * 1000);
	          const hh = String(date.getHours()).padStart(2, "0");
	          return hh;
	        }
	      },
	      grid: {
	        vertLines: { color: "rgba(255,255,255,0.06)" },
	        horzLines: { color: "rgba(255,255,255,0.06)" }
	      },
	      crosshair: {
	        mode: 1,
	        vertLine: {
	          visible: true,
	          labelVisible: false,
	          width: 1,
	          color: "rgba(255,255,255,0.22)",
	          style: 3 // large dashed (closer to TradingView default)
	        },
	        horzLine: {
	          visible: true,
	          labelVisible: false,
	          width: 1,
	          color: "rgba(255,255,255,0.22)",
	          style: 3 // large dashed (closer to TradingView default)
	        }
	      },
	      localization: {
	        locale,
	        priceFormatter: (price: number) => {
	          const formatter = new Intl.NumberFormat(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
          const formatted = formatter.format(price);
          // Keep axis labels compact: show currency in header/tooltips instead of every tick label.
          return formatted;
        }
      }
    });

    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor,
      downColor,
      borderUpColor: upColor,
      borderDownColor: downColor,
      wickUpColor: upColor,
      wickDownColor: downColor,
      priceFormat: { type: "price", precision: 2, minMove: 0.01 }
    });

    const volumeSeries = chart.addSeries(HistogramSeries, {
      color: "rgba(255,255,255,0.4)",
      priceFormat: { type: "volume" },
      priceScaleId: "volume"
    });

    volumeSeries.priceScale().applyOptions({
      scaleMargins: { top: 0.78, bottom: 0.02 }
    });

    // Hide the volume scale labels; keep just the bars like TradingView/Binance.
    chart.priceScale("volume").applyOptions({
      visible: false
    });

	    const handleCrosshairMove = (param: MouseEventParams) => {
	      const hoverCb = onHoverRef.current;
	      if (!hoverCb) return;
	      const tooltipEl = timeTooltipRef.current;
	      if (!param.time || !isUTCTimestamp(param.time)) {
	        hoverCb(null);
	        tooltipEl?.classList.add("hidden");
	        return;
	      }

	      const timeMs = Number(param.time) * 1000;

	      // X-axis hover tooltip: show full date+time above the time scale.
	      if (tooltipEl && param.point) {
	        tooltipEl.textContent = hoverTimeFormatter.format(new Date(timeMs));
	        const width = el.clientWidth || 0;
	        const margin = 40;
	        const clampedX = Math.max(margin, Math.min(width - margin, param.point.x));
	        tooltipEl.style.left = `${clampedX}px`;
	        tooltipEl.classList.remove("hidden");
	      }

	      const candleData = param.seriesData.get(candleSeries) as SeriesDataItemTypeMap["Candlestick"] | undefined;
	      const volumeData = param.seriesData.get(volumeSeries) as SeriesDataItemTypeMap["Histogram"] | undefined;
	      if (!candleData || typeof (candleData as any).open !== "number") {
	        hoverCb(null);
	        return;
	      }

	      const volumeValue =
	        volumeData && typeof (volumeData as any).value === "number" ? ((volumeData as any).value as number) : null;

	      hoverCb({
	        timeMs,
	        open: (candleData as any).open as number,
	        high: (candleData as any).high as number,
	        low: (candleData as any).low as number,
	        close: (candleData as any).close as number,
	        volume: volumeValue
	      });
	    };

	    chart.subscribeCrosshairMove(handleCrosshairMove);

	    const handleMouseLeave = () => {
	      timeTooltipRef.current?.classList.add("hidden");
	    };
	    el.addEventListener("mouseleave", handleMouseLeave);

    const handleVisibleLogicalRangeChange = (range: IRange<number> | null) => {
      visibleLogicalRangeRef.current = range;

      const cb = onViewChangeRef.current;
      if (!cb) return;

      const timeScale = chart.timeScale();
      const barSpacing = Number(timeScale.options().barSpacing ?? 10);
      const scrollPosition = Number(timeScale.scrollPosition());

      pendingViewRef.current = {
        barSpacing: Number.isFinite(barSpacing) ? barSpacing : 10,
        scrollPosition: Number.isFinite(scrollPosition) ? scrollPosition : 0
      };

      if (viewEmitTimerRef.current !== null) return;
      viewEmitTimerRef.current = window.setTimeout(() => {
        viewEmitTimerRef.current = null;
        const payload = pendingViewRef.current;
        pendingViewRef.current = null;
        if (!payload) return;
        onViewChangeRef.current?.(payload);
      }, 200);
    };
    chart.timeScale().subscribeVisibleLogicalRangeChange(handleVisibleLogicalRangeChange);

    if (typeof initialView.scrollPosition === "number" && Number.isFinite(initialView.scrollPosition)) {
      try {
        chart.timeScale().scrollToPosition(initialView.scrollPosition, false);
      } catch {
        // ignore
      }
    }

    chartRef.current = chart;
    candleSeriesRef.current = candleSeries;
    volumeSeriesRef.current = volumeSeries;

	    return () => {
	      if (viewEmitTimerRef.current !== null) {
	        window.clearTimeout(viewEmitTimerRef.current);
	        viewEmitTimerRef.current = null;
	      }
	      el.removeEventListener("mouseleave", handleMouseLeave);
	      chart.unsubscribeCrosshairMove(handleCrosshairMove);
	      chart.timeScale().unsubscribeVisibleLogicalRangeChange(handleVisibleLogicalRangeChange);
	      chart.remove();
	      chartRef.current = null;
	      candleSeriesRef.current = null;
      volumeSeriesRef.current = null;
    };
  }, [locale]);

  const candleData = useMemo<CandlestickData[]>(
    () =>
      bars.map((bar) => ({
        time: bar.timeSec,
        open: bar.open,
        high: bar.high,
        low: bar.low,
        close: bar.close
      })),
    [bars]
  );

  const volumeData = useMemo<HistogramData[]>(
    () =>
      bars.map((bar) => ({
        time: bar.timeSec,
        value: bar.volume,
        color: bar.close >= bar.open ? "rgba(14,203,129,0.45)" : "rgba(246,70,93,0.45)"
      })),
    [bars]
  );

  useEffect(() => {
    const candleSeries = candleSeriesRef.current;
    const volumeSeries = volumeSeriesRef.current;
    const chart = chartRef.current;
    if (!candleSeries || !volumeSeries || !chart) return;

    const timeScale = chart.timeScale();
    const prevVisibleLogicalRange = visibleLogicalRangeRef.current ?? timeScale.getVisibleLogicalRange();
    const wasAtRealTime = Math.abs(timeScale.scrollPosition()) <= 0.5;

    candleSeries.setData(candleData);
    volumeSeries.setData(volumeData);

    // Keep the same zoom/viewport across timeframe changes:
    // - If the user is at the right edge, stay in realtime follow mode.
    // - Otherwise restore the previous visible time range.
    if (!viewInitializedRef.current) {
      viewInitializedRef.current = true;
      if (candleData.length > 0) {
        const initialView = initialViewRef.current;
        const hasInitialBarSpacing = typeof initialView.barSpacing === "number" && Number.isFinite(initialView.barSpacing);
        const hasInitialScrollPosition = typeof initialView.scrollPosition === "number" && Number.isFinite(initialView.scrollPosition);

        if (!hasInitialBarSpacing) {
          timeScale.fitContent();
        } else {
          try {
            timeScale.applyOptions({ barSpacing: initialView.barSpacing! });
          } catch {
            // ignore
          }
        }
        if (hasInitialScrollPosition) {
          try {
            timeScale.scrollToPosition(initialView.scrollPosition!, false);
          } catch {
            // ignore
          }
        } else {
          timeScale.scrollToRealTime();
        }
      }
      return;
    }

    if (candleData.length === 0) {
      return;
    }

    if (wasAtRealTime) {
      timeScale.scrollToRealTime();
      return;
    }

    if (prevVisibleLogicalRange) {
      try {
        timeScale.setVisibleLogicalRange(prevVisibleLogicalRange);
      } catch {
        // ignore
      }
    }
  }, [candleData, volumeData]);

  useEffect(() => {
    if (!liveBar) return;
    const candleSeries = candleSeriesRef.current;
    const volumeSeries = volumeSeriesRef.current;
    if (!candleSeries || !volumeSeries) return;

    candleSeries.update({
      time: liveBar.timeSec,
      open: liveBar.open,
      high: liveBar.high,
      low: liveBar.low,
      close: liveBar.close
    });

    volumeSeries.update({
      time: liveBar.timeSec,
      value: liveBar.volume,
      color: liveBar.close >= liveBar.open ? "rgba(14,203,129,0.45)" : "rgba(246,70,93,0.45)"
    });
  }, [liveBar]);

  return (
    <div style={{ width: "100%", height }} className="relative" data-testid="tv-chart-wrap">
      <div ref={containerRef} style={{ width: "100%", height: "100%" }} data-testid="tv-chart" />
      <div
        ref={timeTooltipRef}
        className="pointer-events-none absolute bottom-2 z-20 hidden -translate-x-1/2 rounded-sm border border-white/10 bg-black/70 px-2 py-0.5 font-mono text-[11px] text-foreground backdrop-blur"
        style={{ left: "50%" }}
      />
    </div>
  );
}
