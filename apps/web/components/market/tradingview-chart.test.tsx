import React from "react";
import { act, cleanup, render } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createChart } from "lightweight-charts";
import { TradingViewChart } from "./tradingview-chart";

let crosshairHandler: ((param: any) => void) | null = null;
const timeScaleApi = {
  options: () => ({ barSpacing: 10 }),
  scrollPosition: () => 0,
  applyOptions: vi.fn(),
  setVisibleLogicalRange: vi.fn(),
  subscribeVisibleLogicalRangeChange: vi.fn(),
  unsubscribeVisibleLogicalRangeChange: vi.fn(),
  scrollToPosition: vi.fn(),
  fitContent: vi.fn()
};

const rightPriceScaleApi = {
  options: () => ({ autoScale: true }),
  getVisibleRange: vi.fn(() => null),
  setAutoScale: vi.fn(),
  setVisibleRange: vi.fn(),
  applyOptions: vi.fn()
};

const leftPriceScaleApi = {
  options: () => ({ autoScale: true }),
  getVisibleRange: vi.fn(() => null),
  setAutoScale: vi.fn(),
  setVisibleRange: vi.fn(),
  applyOptions: vi.fn()
};

const volumePriceScaleApi = {
  applyOptions: vi.fn()
};

const chartApplyOptionsMock = vi.fn();
const seriesInstances: Array<{ setData: ReturnType<typeof vi.fn>; update: ReturnType<typeof vi.fn>; applyOptions: ReturnType<typeof vi.fn> }> = [];

function createSeriesMock() {
  const instance = {
    setData: vi.fn(),
    update: vi.fn(),
    applyOptions: vi.fn()
  };
  seriesInstances.push(instance);
  return instance;
}

vi.mock("lightweight-charts", () => {
  return {
    AreaSeries: {},
    BarSeries: {},
    CandlestickSeries: {},
    HistogramSeries: {},
    LineSeries: {},
    PriceScaleMode: {
      Normal: 0,
      Logarithmic: 1,
      Percentage: 2,
      IndexedTo100: 3
    },
    ColorType: { Solid: "solid" },
    createChart: vi.fn(() => ({
      addSeries: vi.fn(() => createSeriesMock()),
      subscribeCrosshairMove: vi.fn((handler: (param: any) => void) => {
        crosshairHandler = handler;
      }),
      unsubscribeCrosshairMove: vi.fn(),
      timeScale: vi.fn(() => timeScaleApi),
      priceScale: vi.fn((id: string) => {
        if (id === "left") return leftPriceScaleApi;
        if (id === "right") return rightPriceScaleApi;
        return volumePriceScaleApi;
      }),
      applyOptions: chartApplyOptionsMock,
      remove: vi.fn()
    }))
  };
});

describe("TradingViewChart", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    crosshairHandler = null;
    seriesInstances.length = 0;
  });
  afterEach(() => {
    cleanup();
  });

  it("sets initial logical range with past and future padding", () => {
    render(
      <TradingViewChart
        quoteCurrency="USD"
        bars={[
          {
            timeSec: 1739662800 as any,
            open: 100,
            high: 103,
            low: 99,
            close: 102,
            volume: 12
          }
        ]}
      />
    );

    expect(timeScaleApi.setVisibleLogicalRange).toHaveBeenCalledWith({ from: -20, to: 12 });
  });

  it("disables the chart attribution logo", () => {
    render(
      <TradingViewChart
        quoteCurrency="USD"
        bars={[
          {
            timeSec: 1739662800 as any,
            open: 100,
            high: 103,
            low: 99,
            close: 102,
            volume: 12
          }
        ]}
      />
    );

    expect(vi.mocked(createChart)).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        layout: expect.objectContaining({
          attributionLogo: false
        })
      })
    );
  });

  it("does not override logical range when initial scroll position is provided", () => {
    render(
      <TradingViewChart
        quoteCurrency="USD"
        initialScrollPosition={5}
        bars={[
          {
            timeSec: 1739662800 as any,
            open: 100,
            high: 103,
            low: 99,
            close: 102,
            volume: 12
          }
        ]}
      />
    );

    expect(timeScaleApi.setVisibleLogicalRange).not.toHaveBeenCalled();
  });

  it("emits volume on hover from matching bar data", () => {
    const onHover = vi.fn();
    render(
      <TradingViewChart
        quoteCurrency="USD"
        bars={[
          {
            timeSec: 1739662800 as any,
            open: 100,
            high: 103,
            low: 99,
            close: 102,
            volume: 12
          }
        ]}
        onHover={onHover}
      />
    );

    act(() => {
      const activeSeries = seriesInstances[0];
      crosshairHandler?.({
        time: 1739662800,
        point: { x: 64, y: 20 },
        seriesData: new Map([[activeSeries, { open: 100, high: 103, low: 99, close: 102 }]])
      });
    });

    expect(onHover).toHaveBeenLastCalledWith({
      timeMs: 1739662800000,
      open: 100,
      high: 103,
      low: 99,
      close: 102,
      volume: 12
    });
  });

  it("clears volume series data when volume is hidden", () => {
    render(
      <TradingViewChart
        quoteCurrency="USD"
        showVolume={false}
        bars={[
          {
            timeSec: 1739662800 as any,
            open: 100,
            high: 103,
            low: 99,
            close: 102,
            volume: 12
          }
        ]}
      />
    );

    const volumeSeries = seriesInstances[4];
    expect(volumeSeries.setData).toHaveBeenCalledWith([]);
  });

});
