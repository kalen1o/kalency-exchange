"use client";

import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  AreaSeries,
  BarSeries,
  CandlestickSeries,
  HistogramSeries,
  LineSeries,
  createChart,
  ColorType,
  type IChartApi,
  type IRange,
  type ISeriesApi,
  type UTCTimestamp,
  type LineData,
  type BarData,
  type CandlestickData,
  type HistogramData,
  type MouseEventParams
} from "lightweight-charts";
import { debounce } from "@/lib/debounce";

type TimeSec = UTCTimestamp;

enum CandleColors {
  Up = "#0ECB81",
  Down = "#F6465D"
}

const INITIAL_LOGICAL_RANGE_PAST_BARS = 20;
const INITIAL_LOGICAL_RANGE_FUTURE_BARS = 12;

export type TradingViewBar = {
  timeSec: TimeSec;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
};

export type ChartBarStyle = "candles" | "bars" | "line" | "line-area";

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
  showVolume?: boolean;
  onHover?: (hover: TradingViewHover | null) => void;
  initialBarSpacing?: number;
  initialScrollPosition?: number;
  initialPriceFrom?: number;
  initialPriceTo?: number;
  barStyle?: ChartBarStyle;
  onViewChange?: (view: { barSpacing: number; scrollPosition: number; priceFrom: number | null; priceTo: number | null }) => void;
};

export function TradingViewChart({
  height = "100%",
  quoteCurrency: _quoteCurrency,
  bars,
  liveBar,
  showVolume = true,
  onHover,
  initialBarSpacing,
  initialScrollPosition,
  initialPriceFrom,
  initialPriceTo,
  barStyle = "candles",
  onViewChange
}: TradingViewChartProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const timeTooltipRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candleSeriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const barSeriesRef = useRef<ISeriesApi<"Bar"> | null>(null);
  const lineSeriesRef = useRef<ISeriesApi<"Line"> | null>(null);
  const areaSeriesRef = useRef<ISeriesApi<"Area"> | null>(null);
  const volumeSeriesRef = useRef<ISeriesApi<"Histogram"> | null>(null);
  const barStyleRef = useRef<ChartBarStyle>(barStyle);
  const barsVolumeByTimeRef = useRef<Map<number, number>>(new Map());
  const liveBarRef = useRef<TradingViewBar | null | undefined>(liveBar);
  const onHoverRef = useRef<TradingViewChartProps["onHover"]>(onHover);
  const onViewChangeRef = useRef<TradingViewChartProps["onViewChange"]>(onViewChange);
  const didFitContentRef = useRef(false);
  const didInitViewRef = useRef(false);
  const initialBarSpacingRef = useRef(initialBarSpacing);
  const initialScrollPositionRef = useRef(initialScrollPosition);
  const initialPriceFromRef = useRef(initialPriceFrom);
  const initialPriceToRef = useRef(initialPriceTo);

  const upColor = CandleColors.Up;
  const downColor = CandleColors.Down;

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
    barStyleRef.current = barStyle;
  }, [barStyle]);

  useEffect(() => {
    const volumeByTime = new Map<number, number>();
    for (const bar of bars) {
      volumeByTime.set(Number(bar.timeSec), bar.volume);
    }
    barsVolumeByTimeRef.current = volumeByTime;
  }, [bars]);

  useEffect(() => {
    liveBarRef.current = liveBar;
  }, [liveBar]);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    const initialWidth = Math.max(1, el.clientWidth);
    const initialHeight = Math.max(1, el.clientHeight);
    const chart = createChart(el, {
      width: initialWidth,
      height: initialHeight,
      layout: {
        background: { type: ColorType.Solid, color: "#0B0E11" },
        textColor: "#C9D1D9",
        fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
        attributionLogo: false
      },
      rightPriceScale: {
        visible: true,
        borderColor: "rgba(255,255,255,0.10)",
        ticksVisible: true,
        entireTextOnly: false,
        scaleMargins: { top: 0.08, bottom: 0.28 }
      },
      leftPriceScale: {
        visible: false,
        borderColor: "rgba(255,255,255,0.10)",
        ticksVisible: true,
        entireTextOnly: false,
        scaleMargins: { top: 0.08, bottom: 0.28 }
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
      },
      timeScale: {
        borderColor: "rgba(255,255,255,0.10)"
      }
    });

    if (typeof initialBarSpacingRef.current === "number" && Number.isFinite(initialBarSpacingRef.current)) {
      chart.timeScale().applyOptions({ barSpacing: initialBarSpacingRef.current });
    }

    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor,
      downColor,
      borderUpColor: upColor,
      borderDownColor: downColor,
      wickUpColor: upColor,
      wickDownColor: downColor,
      priceFormat: { type: "price", precision: 2, minMove: 0.01 }
    });
    const barSeries = chart.addSeries(BarSeries, {
      upColor,
      downColor,
      thinBars: false,
      openVisible: true,
      priceFormat: { type: "price", precision: 2, minMove: 0.01 }
    });
    const lineSeries = chart.addSeries(LineSeries, {
      color: "#EAB308",
      lineWidth: 2,
      crosshairMarkerVisible: false,
      priceFormat: { type: "price", precision: 2, minMove: 0.01 }
    });
    const areaSeries = chart.addSeries(AreaSeries, {
      lineColor: "#38BDF8",
      topColor: "rgba(56, 189, 248, 0.35)",
      bottomColor: "rgba(56, 189, 248, 0.05)",
      lineWidth: 2,
      crosshairMarkerVisible: false,
      priceFormat: { type: "price", precision: 2, minMove: 0.01 }
    });
    candleSeries.applyOptions({ visible: barStyleRef.current === "candles" });
    barSeries.applyOptions({ visible: barStyleRef.current === "bars" });
    lineSeries.applyOptions({ visible: barStyleRef.current === "line" });
    areaSeries.applyOptions({ visible: barStyleRef.current === "line-area" });
    const volumeSeries = chart.addSeries(HistogramSeries, {
      priceScaleId: "",
      priceFormat: { type: "volume" }
    });
    chart.priceScale("").applyOptions({
      scaleMargins: { top: 0.75, bottom: 0 }
    });

    const readScaleApi = (side: PriceAxisSide) =>
      chart.priceScale(side) as {
        options?: () => { autoScale?: boolean };
        getVisibleRange?: () => { from: number; to: number } | null;
        setAutoScale?: (autoScale: boolean) => void;
        setVisibleRange?: (range: { from: number; to: number }) => void;
      };
    const activePriceScale = () => readScaleApi("right");

    if (
      typeof initialPriceFromRef.current === "number" &&
      Number.isFinite(initialPriceFromRef.current) &&
      typeof initialPriceToRef.current === "number" &&
      Number.isFinite(initialPriceToRef.current)
    ) {
      const scale = activePriceScale();
      scale.setAutoScale?.(false);
      scale.setVisibleRange?.({
        from: initialPriceFromRef.current,
        to: initialPriceToRef.current
      });
    }

    const handleCrosshairMove = (param: MouseEventParams) => {
      const hoverCb = onHoverRef.current;
      if (!hoverCb) return;
      const tooltipEl = timeTooltipRef.current;
      if (!param.time || typeof param.time !== "number") {
        hoverCb(null);
        tooltipEl?.classList.add("hidden");
        return;
      }

      const timeMs = param.time * 1000;
      if (tooltipEl && param.point) {
        tooltipEl.textContent = hoverTimeFormatter.format(new Date(timeMs));
        const width = el.clientWidth || 0;
        const margin = 40;
        const clampedX = Math.max(margin, Math.min(width - margin, param.point.x));
        tooltipEl.style.left = `${clampedX}px`;
        tooltipEl.classList.remove("hidden");
      }

      const activeSeries =
        barStyleRef.current === "bars"
          ? barSeries
          : barStyleRef.current === "line"
            ? lineSeries
            : barStyleRef.current === "line-area"
              ? areaSeries
              : candleSeries;
      const candleData = param.seriesData.get(activeSeries) as any;
      if (!candleData) {
        hoverCb(null);
        return;
      }
      const seriesValue = Number(candleData.value);
      const openValue = Number(candleData.open);
      const highValue = Number(candleData.high);
      const lowValue = Number(candleData.low);
      const closeValue = Number(candleData.close);
      const isLineLike = barStyleRef.current === "line" || barStyleRef.current === "line-area";
      if (isLineLike && !Number.isFinite(seriesValue)) {
        hoverCb(null);
        return;
      }
      if (!isLineLike && (!Number.isFinite(openValue) || !Number.isFinite(highValue) || !Number.isFinite(lowValue) || !Number.isFinite(closeValue))) {
        hoverCb(null);
        return;
      }

      hoverCb({
        timeMs,
        open: isLineLike ? seriesValue : openValue,
        high: isLineLike ? seriesValue : highValue,
        low: isLineLike ? seriesValue : lowValue,
        close: isLineLike ? seriesValue : closeValue,
        volume: (() => {
          const timeSec = Number(param.time);
          const currentLiveBar = liveBarRef.current;
          if (
            currentLiveBar &&
            Number(currentLiveBar.timeSec) === timeSec &&
            Number.isFinite(currentLiveBar.volume)
          ) {
            return currentLiveBar.volume;
          }
          const matchedVolume = barsVolumeByTimeRef.current.get(timeSec);
          return typeof matchedVolume === "number" && Number.isFinite(matchedVolume) ? matchedVolume : null;
        })()
      });
    };

    chart.subscribeCrosshairMove(handleCrosshairMove);

    const handleMouseLeave = () => {
      timeTooltipRef.current?.classList.add("hidden");
      onHoverRef.current?.(null);
    };
    el.addEventListener("mouseleave", handleMouseLeave);

    const emitViewChange = debounce((view: { barSpacing: number; scrollPosition: number; priceFrom: number | null; priceTo: number | null }) => {
      onViewChangeRef.current?.(view);
    }, 250);

    const readPriceView = () => {
      const scale = activePriceScale();
      const autoScale = scale.options?.()?.autoScale ?? true;
      if (autoScale) {
        return { priceFrom: null, priceTo: null };
      }
      const visible = scale.getVisibleRange?.();
      if (!visible) {
        return { priceFrom: null, priceTo: null };
      }
      const from = Number(visible.from);
      const to = Number(visible.to);
      if (!Number.isFinite(from) || !Number.isFinite(to)) {
        return { priceFrom: null, priceTo: null };
      }
      return { priceFrom: from, priceTo: to };
    };

    const emitCurrentView = () => {
      const cb = onViewChangeRef.current;
      if (!cb) return;
      const timeScale = chart.timeScale();
      const barSpacing = Number(timeScale.options().barSpacing ?? 10);
      const scrollPosition = Number(timeScale.scrollPosition());
      const priceView = readPriceView();
      emitViewChange({
        barSpacing: Number.isFinite(barSpacing) ? barSpacing : 10,
        scrollPosition: Number.isFinite(scrollPosition) ? scrollPosition : 0,
        priceFrom: priceView.priceFrom,
        priceTo: priceView.priceTo
      });
    };

    const handleVisibleLogicalRangeChange = (_range: IRange<number> | null) => {
      emitCurrentView();
    };
    chart.timeScale().subscribeVisibleLogicalRangeChange(handleVisibleLogicalRangeChange);

    const handleAxisInteractionEnd = () => {
      window.setTimeout(() => {
        emitCurrentView();
      }, 0);
    };
    el.addEventListener("pointerup", handleAxisInteractionEnd);
    el.addEventListener("wheel", handleAxisInteractionEnd, { passive: true });
    el.addEventListener("dblclick", handleAxisInteractionEnd);

    chartRef.current = chart;
    candleSeriesRef.current = candleSeries;
    barSeriesRef.current = barSeries;
    lineSeriesRef.current = lineSeries;
    areaSeriesRef.current = areaSeries;
    volumeSeriesRef.current = volumeSeries;
    didFitContentRef.current = false;
    didInitViewRef.current = false;

    const handleResize = () => {
      const width = Math.max(1, el.clientWidth);
      const height = Math.max(1, el.clientHeight);
      chart.applyOptions({ width, height });
    };

    window.addEventListener("resize", handleResize);
    handleResize();

    return () => {
      window.removeEventListener("resize", handleResize);
      el.removeEventListener("mouseleave", handleMouseLeave);
      el.removeEventListener("pointerup", handleAxisInteractionEnd);
      el.removeEventListener("wheel", handleAxisInteractionEnd);
      el.removeEventListener("dblclick", handleAxisInteractionEnd);
      chart.unsubscribeCrosshairMove(handleCrosshairMove);
      chart.timeScale().unsubscribeVisibleLogicalRangeChange(handleVisibleLogicalRangeChange);
      emitViewChange.flush();
      emitViewChange.cancel();
      chart.remove();
      chartRef.current = null;
      candleSeriesRef.current = null;
      barSeriesRef.current = null;
      lineSeriesRef.current = null;
      areaSeriesRef.current = null;
      volumeSeriesRef.current = null;
    };
  }, [locale]);

  const candleData = useMemo<ReadonlyArray<CandlestickData>>(() => {
    return bars.map((bar) => ({
      time: bar.timeSec,
      open: bar.open,
      high: bar.high,
      low: bar.low,
      close: bar.close
    }));
  }, [bars]);
  const volumeData = useMemo<ReadonlyArray<HistogramData>>(() => {
    return bars.map((bar) => ({
      time: bar.timeSec,
      value: bar.volume,
      color: bar.close >= bar.open ? "rgba(14, 203, 129, 0.4)" : "rgba(246, 70, 93, 0.4)"
    }));
  }, [bars]);
  const barData = useMemo<ReadonlyArray<BarData>>(() => {
    return bars.map((bar) => ({
      time: bar.timeSec,
      open: bar.open,
      high: bar.high,
      low: bar.low,
      close: bar.close
    }));
  }, [bars]);
  const lineData = useMemo<ReadonlyArray<LineData>>(() => {
    return bars.map((bar) => ({
      time: bar.timeSec,
      value: bar.close
    }));
  }, [bars]);

  useEffect(() => {
    const candleSeries = candleSeriesRef.current;
    const barSeries = barSeriesRef.current;
    const lineSeries = lineSeriesRef.current;
    const areaSeries = areaSeriesRef.current;
    const volumeSeries = volumeSeriesRef.current;
    const chart = chartRef.current;
    if (!candleSeries || !barSeries || !lineSeries || !areaSeries || !volumeSeries || !chart) return;
    candleSeries.setData(candleData);
    barSeries.setData(barData);
    lineSeries.setData(lineData);
    areaSeries.setData(lineData);
    candleSeries.applyOptions({ visible: barStyle === "candles" });
    barSeries.applyOptions({ visible: barStyle === "bars" });
    lineSeries.applyOptions({ visible: barStyle === "line" });
    areaSeries.applyOptions({ visible: barStyle === "line-area" });
    volumeSeries.setData(showVolume ? volumeData : []);

    if (!didInitViewRef.current && candleData.length > 0) {
      didInitViewRef.current = true;
      if (
        typeof initialScrollPositionRef.current === "number" &&
        Number.isFinite(initialScrollPositionRef.current)
      ) {
        chart.timeScale().scrollToPosition(initialScrollPositionRef.current, false);
        didFitContentRef.current = true;
        return;
      }
    }

    if (!didFitContentRef.current && candleData.length > 0) {
      didFitContentRef.current = true;
      chart.timeScale().setVisibleLogicalRange({
        from: -INITIAL_LOGICAL_RANGE_PAST_BARS,
        to: candleData.length - 1 + INITIAL_LOGICAL_RANGE_FUTURE_BARS
      });
    }
  }, [candleData, barData, lineData, volumeData, showVolume, barStyle]);

  useEffect(() => {
    if (!liveBar) return;
    const candleSeries = candleSeriesRef.current;
    const barSeries = barSeriesRef.current;
    const lineSeries = lineSeriesRef.current;
    const areaSeries = areaSeriesRef.current;
    const volumeSeries = volumeSeriesRef.current;
    if (!candleSeries || !barSeries || !lineSeries || !areaSeries || !volumeSeries) return;

    candleSeries.update({
      time: liveBar.timeSec,
      open: liveBar.open,
      high: liveBar.high,
      low: liveBar.low,
      close: liveBar.close
    });
    barSeries.update({
      time: liveBar.timeSec,
      open: liveBar.open,
      high: liveBar.high,
      low: liveBar.low,
      close: liveBar.close
    });
    lineSeries.update({
      time: liveBar.timeSec,
      value: liveBar.close
    });
    areaSeries.update({
      time: liveBar.timeSec,
      value: liveBar.close
    });
    if (!showVolume) return;
    volumeSeries.update({
      time: liveBar.timeSec,
      value: liveBar.volume,
      color: liveBar.close >= liveBar.open ? "rgba(14, 203, 129, 0.4)" : "rgba(246, 70, 93, 0.4)"
    });
  }, [liveBar, showVolume]);

  return (
    <div style={{ width: "100%", height }} className="relative" data-testid="tv-chart-wrap">
      <div ref={containerRef} style={{ width: "100%", height: "100%" }} data-testid="tv-chart" />
      <div
        ref={timeTooltipRef}
        className="pointer-events-none absolute bottom-8 z-20 hidden -translate-x-1/2 rounded-sm border border-white/10 bg-black/70 px-2 py-0.5 font-mono text-[11px] text-foreground backdrop-blur"
        style={{ left: "50%" }}
      />
    </div>
  );
}
