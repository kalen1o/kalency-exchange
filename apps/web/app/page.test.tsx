import React from "react";
import { cleanup, fireEvent, render, screen, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { withNuqsTestingAdapter } from "nuqs/adapters/testing";
import { formatLocaleTime } from "@/lib/datetime";
import HomeClient from "./home-client";

vi.mock("@/components/market/tradingview-chart", () => {
  return {
    TradingViewChart: ({ onHover }: { onHover?: (hover: any) => void }) => (
      <div
        data-testid="tv-chart-mock"
        onMouseMove={() => {
          onHover?.({
            timeMs: Date.parse("2026-02-15T00:01:00Z"),
            open: 100.5,
            high: 103,
            low: 100,
            close: 102.2,
            volume: 12
          });
        }}
        onMouseLeave={() => {
          onHover?.(null);
        }}
      />
    )
  };
});

describe("Home trading workspace layout", () => {
  const wsMock = vi.fn();
  const sampleTrades = [
    {
      tradeId: "trade-1",
      symbol: "BTC-USD",
      price: 102.25,
      qty: 1.2,
      ts: "2026-02-15T00:00:10Z"
    }
  ];
  const sampleCandles = [
    {
      symbol: "BTC-USD",
      timeframe: "1m",
      bucketStart: "2026-02-15T00:00:00Z",
      open: 100,
      high: 101,
      low: 99,
      close: 100.5,
      volume: 10
    },
    {
      symbol: "BTC-USD",
      timeframe: "1m",
      bucketStart: "2026-02-15T00:01:00Z",
      open: 100.5,
      high: 103,
      low: 100,
      close: 102.2,
      volume: 12
    }
  ];

  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input) => {
        const url = String(input);
        if (url.includes("/v1/orders/open")) {
          return new Response(JSON.stringify([]), { status: 200 });
        }
        if (url.includes("/candles")) {
          return new Response(JSON.stringify(sampleCandles), { status: 200 });
        }
        if (url.includes("/trades")) {
          return new Response(JSON.stringify(sampleTrades), { status: 200 });
        }
        return new Response(JSON.stringify([]), { status: 200 });
      })
    );

    class MockWebSocket {
      url: string;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onclose: (() => void) | null = null;
      onerror: (() => void) | null = null;

      constructor(url: string) {
        this.url = url;
        wsMock(url);
      }

      close() {
        return undefined;
      }
    }

    vi.stubGlobal("WebSocket", MockWebSocket as unknown as typeof WebSocket);
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    wsMock.mockReset();
  });

  it("renders chart workspace with pairs sidebar and user-panel modal", async () => {
    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });

    const workspace = await screen.findByTestId("workspace-grid");
    const chartPanel = within(workspace).getByTestId("chart-panel");
    const pairsSidebar = within(workspace).getByTestId("pairs-sidebar");

    expect(chartPanel).toBeInTheDocument();
    expect(pairsSidebar).toBeInTheDocument();
    expect(within(pairsSidebar).getByRole("button", { name: /btc-usd/i })).toBeInTheDocument();
    expect(within(pairsSidebar).getByRole("button", { name: /eth-usd/i })).toBeInTheDocument();
    expect(screen.getByTestId("user-avatar")).toBeInTheDocument();
    expect(screen.getByTestId("chart-header-price")).toHaveTextContent("Live: 102.25 USD");
    expect(screen.getByTestId("chart-live-change")).toHaveTextContent("Change: +1.70 USD (+1.69%)");
    expect(screen.getByTestId("chart-live-open")).toHaveTextContent("Open: 100.50 USD");
    expect(screen.getByTestId("chart-live-open")).toHaveClass("text-emerald-400");
    expect(screen.getByTestId("chart-live-close")).toHaveTextContent("Close: 102.20 USD");
    expect(screen.getByTestId("chart-live-close")).toHaveClass("text-rose-400");
    expect(screen.queryByText(/O 100\.50 USD \| C 102\.20 USD/)).not.toBeInTheDocument();
    expect(screen.queryByText(/^V 12$/)).not.toBeInTheDocument();
    expect(screen.getByTestId("tv-chart-mock")).toBeInTheDocument();

    fireEvent.click(screen.getByTestId("user-avatar"));
    expect(screen.getByTestId("user-menu")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /order ticket/i }));
    expect(await screen.findByTestId("user-panel")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /submit order/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/quantity/i)).toBeInTheDocument();
  });

  it("opens ticks websocket for live updates", async () => {
    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });

    const workspaces = await screen.findAllByTestId("workspace-grid");
    expect(workspaces.length).toBeGreaterThan(0);
    expect(wsMock).toHaveBeenCalledWith("ws://localhost:8080/ws/ticks/BTC-USD");
  });

  it("shows hovered candle values when hovering the chart", async () => {
    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });

    const headerPrice = await screen.findByTestId("chart-header-price");
    const chart = screen.getByTestId("tv-chart-mock");

    fireEvent.mouseMove(chart);
    expect(headerPrice).toHaveTextContent(`Hover: 102.20 USD @ ${formatLocaleTime("2026-02-15T00:01:00Z")}`);
    expect(screen.getByTestId("chart-live-change")).toHaveTextContent("Change: +1.70 USD (+1.69%)");
    expect(screen.getByTestId("chart-live-open")).toHaveTextContent("Open: 100.50 USD");
    expect(screen.getByTestId("chart-live-close")).toHaveTextContent("Close: 102.20 USD");

    fireEvent.mouseLeave(chart);
    expect(headerPrice).toHaveTextContent("Live: 102.25 USD");
    expect(screen.getByTestId("chart-live-change")).toHaveTextContent("Change: +1.70 USD (+1.69%)");
    expect(screen.getByTestId("chart-live-open")).toHaveTextContent("Open: 100.50 USD");
    expect(screen.getByTestId("chart-live-close")).toHaveTextContent("Close: 102.20 USD");
  });
});
