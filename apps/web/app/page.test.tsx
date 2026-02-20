import React from "react";
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { withNuqsTestingAdapter } from "nuqs/adapters/testing";
import HomeClient from "./home-client";

vi.mock("@/components/market/tradingview-chart", () => {
  return {
    TradingViewChart: ({ onHover, showVolume = true }: { onHover?: (hover: any) => void; showVolume?: boolean }) => (
      <div
        data-testid="tv-chart-mock"
        data-show-volume={showVolume ? "true" : "false"}
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
      makerUserId: "sim-maker-BTC-USD",
      takerUserId: "sim-taker-BTC-USD",
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
    expect(workspace.className).toContain("overflow-hidden");
    expect(pairsSidebar).toBeInTheDocument();
    expect(pairsSidebar.className).toContain("overflow-y-auto");
    expect(within(pairsSidebar).getByRole("button", { name: /btc-usd/i })).toBeInTheDocument();
    expect(within(pairsSidebar).getByRole("button", { name: /eth-usd/i })).toBeInTheDocument();
    expect(within(pairsSidebar).getByRole("button", { name: /btc-eth/i })).toBeInTheDocument();
    expect(screen.getByTestId("user-avatar")).toBeInTheDocument();
    expect(screen.getByTestId("chart-header-price")).toHaveTextContent("102.25 USD");
    expect(screen.getByTestId("chart-live-open")).toHaveTextContent("O: 100.50 USD");
    expect(screen.getByTestId("chart-live-open")).toHaveClass("text-emerald-400");
    expect(screen.getByTestId("chart-live-close")).toHaveTextContent("C: 102.20 USD");
    expect(screen.getByTestId("chart-live-close")).toHaveClass("text-emerald-400");
    expect(screen.getByTestId("chart-live-high")).toHaveTextContent("H: 103.00 USD");
    expect(screen.getByTestId("chart-live-high")).toHaveClass("text-emerald-400");
    expect(screen.getByTestId("chart-live-low")).toHaveTextContent("L: 100.00 USD");
    expect(screen.getByTestId("chart-live-low")).toHaveClass("text-emerald-400");
    expect(screen.getByTestId("chart-live-volume")).toHaveTextContent("V: 12");
    expect(screen.getByTestId("chart-live-volume")).toHaveClass("text-emerald-400");
    expect(screen.getByTestId("chart-live-change-pct")).toHaveTextContent("+1.69%");
    expect(screen.getByTestId("tv-chart-mock")).toHaveAttribute("data-show-volume", "true");
    expect(screen.queryByText(/O 100\.50 USD \| C 102\.20 USD/)).not.toBeInTheDocument();
    expect(screen.getByTestId("tv-chart-mock")).toBeInTheDocument();
    expect(screen.getByTestId("pairs-open-button")).toBeInTheDocument();

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

  it("shows trade IN/OUT participants in user panel", async () => {
    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });
    await screen.findByTestId("workspace-grid");

    fireEvent.click(screen.getByTestId("header-trades-button"));

    expect(await screen.findByTestId("user-panel")).toBeInTheDocument();
    expect(screen.getByText("IN")).toBeInTheDocument();
    expect(screen.getByText("OUT")).toBeInTheDocument();
    expect(screen.getByText("sim-maker-BTC-USD")).toBeInTheDocument();
    expect(screen.getByText("sim-taker-BTC-USD")).toBeInTheDocument();
  });

  it("shows hovered candle values when hovering the chart", async () => {
    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });

    const headerPrice = await screen.findByTestId("chart-header-price");
    const chart = screen.getByTestId("tv-chart-mock");

    fireEvent.mouseMove(chart);
    expect(headerPrice).toHaveTextContent(/^102.20 USD$/);
    expect(screen.getByTestId("chart-live-open")).toHaveTextContent("O: 100.50 USD");
    expect(screen.getByTestId("chart-live-close")).toHaveTextContent("C: 102.20 USD");
    expect(screen.getByTestId("chart-live-high")).toHaveTextContent("H: 103.00 USD");
    expect(screen.getByTestId("chart-live-low")).toHaveTextContent("L: 100.00 USD");
    expect(screen.getByTestId("chart-live-volume")).toHaveTextContent("V: 12");
    expect(screen.getByTestId("chart-live-change-pct")).toHaveTextContent("+1.69%");

    fireEvent.mouseLeave(chart);
    expect(headerPrice).toHaveTextContent("102.25 USD");
    expect(screen.getByTestId("chart-live-open")).toHaveTextContent("O: 100.50 USD");
    expect(screen.getByTestId("chart-live-close")).toHaveTextContent("C: 102.20 USD");
    expect(screen.getByTestId("chart-live-high")).toHaveTextContent("H: 103.00 USD");
    expect(screen.getByTestId("chart-live-low")).toHaveTextContent("L: 100.00 USD");
    expect(screen.getByTestId("chart-live-volume")).toHaveTextContent("V: 12");
    expect(screen.getByTestId("chart-live-change-pct")).toHaveTextContent("+1.69%");
  });

  it("toggles volume visibility from indicators popup", async () => {
    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });

    const workspace = await screen.findByTestId("workspace-grid");
    const chartPanel = within(workspace).getByTestId("chart-panel");
    const topHeader = screen.getByTestId("top-header");
    expect(screen.getByTestId("chart-live-volume")).toHaveTextContent("V: 12");
    expect(screen.getByTestId("tv-chart-mock")).toHaveAttribute("data-show-volume", "true");
    expect(within(chartPanel).queryByTestId("chart-indicators-button")).not.toBeInTheDocument();
    expect(within(topHeader).getByTestId("chart-indicators-button")).toBeInTheDocument();

    fireEvent.click(within(topHeader).getByTestId("chart-indicators-button"));
    expect(screen.getByTestId("chart-indicators-menu")).toBeInTheDocument();

    const volumeSwitch = screen.getByTestId("chart-indicators-volume-switch");
    expect(volumeSwitch).toHaveAttribute("aria-checked", "true");

    fireEvent.click(volumeSwitch);
    expect(volumeSwitch).toHaveAttribute("aria-checked", "false");
    expect(screen.queryByTestId("chart-live-volume")).not.toBeInTheDocument();
    expect(screen.getByTestId("tv-chart-mock")).toHaveAttribute("data-show-volume", "false");
  });

  it("opens pairs search modal from header", async () => {
    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });

    await screen.findByTestId("workspace-grid");
    fireEvent.click(screen.getByTestId("pairs-open-button"));

    expect(screen.getByTestId("pairs-search-modal")).toBeInTheDocument();
    expect(screen.getByTestId("pairs-search-input")).toBeInTheDocument();
  });

  it("selects searched pair from header modal", async () => {
    const fetchMock = vi.fn(async (input) => {
      const url = String(input);
      if (url.includes("api.binance.com/api/v3/exchangeInfo")) {
        return new Response(
          JSON.stringify({
            symbols: [{ symbol: "SOLUSDT", status: "TRADING", isSpotTradingAllowed: true }]
          }),
          { status: 200 }
        );
      }
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
    });
    vi.stubGlobal(
      "fetch",
      fetchMock
    );

    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });
    await screen.findByTestId("workspace-grid");

    fireEvent.click(screen.getByTestId("pairs-open-button"));
    fireEvent.change(screen.getByTestId("pairs-search-input"), { target: { value: "sol" } });

    const solButton = await screen.findByRole("button", { name: /sol-usd/i });
    fireEvent.click(solButton);

    expect(screen.getByTestId("pairs-open-button")).toHaveTextContent("SOL-USD");
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/v1/admin/symbols/SOL-USD/ensure",
      expect.objectContaining({
        method: "POST"
      })
    );
  });

  it("registers beforeunload confirmation when tab is closing", async () => {
    const addEventListenerSpy = vi.spyOn(window, "addEventListener");
    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });
    await screen.findByTestId("workspace-grid");

    const beforeUnloadCall = addEventListenerSpy.mock.calls.find((call) => call[0] === "beforeunload");
    expect(beforeUnloadCall).toBeDefined();

    const beforeUnloadHandler = beforeUnloadCall?.[1] as (event: Event & { returnValue?: string }) => void;
    const event = new Event("beforeunload", { cancelable: true }) as Event & { returnValue?: string };
    event.returnValue = undefined;
    beforeUnloadHandler(event);

    expect(event.defaultPrevented).toBe(true);
    expect(event.returnValue).toBe(false);
  });

  it("warms up simulation for additional searched pairs on init", async () => {
    const fetchMock = vi.fn(async (input, init?: RequestInit) => {
      const url = String(input);
      if (url.includes("api.binance.com/api/v3/exchangeInfo")) {
        return new Response(
          JSON.stringify({
            symbols: [{ symbol: "SOLUSDT", status: "TRADING", isSpotTradingAllowed: true }]
          }),
          { status: 200 }
        );
      }
      if (url.includes("/v1/orders/open")) {
        return new Response(JSON.stringify([]), { status: 200 });
      }
      if (url.includes("/candles")) {
        return new Response(JSON.stringify(sampleCandles), { status: 200 });
      }
      if (url.includes("/trades")) {
        return new Response(JSON.stringify(sampleTrades), { status: 200 });
      }
      if (url.includes("/v1/admin/symbols/")) {
        const ensuredSymbol = decodeURIComponent(url.split("/v1/admin/symbols/")[1].split("/")[0] ?? "");
        return new Response(JSON.stringify({ symbol: ensuredSymbol, ensured: true }), { status: 200 });
      }
      return new Response(JSON.stringify([]), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<HomeClient />, { wrapper: withNuqsTestingAdapter({ hasMemory: true }) });
    await screen.findByTestId("workspace-grid");

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "http://localhost:8080/v1/admin/symbols/SOL-USD/ensure",
        expect.objectContaining({ method: "POST" })
      );
    });

    expect(screen.getByTestId("pairs-sidebar")).toHaveTextContent("SOL-USD");
  });

  it("normalizes unknown pair query to BTC-USD", async () => {
    render(<HomeClient />, {
      wrapper: withNuqsTestingAdapter({
        hasMemory: true,
        searchParams: "?pair=GAS-BTC"
      })
    });

    await screen.findByTestId("workspace-grid");
    expect(screen.getByTestId("pairs-open-button")).toHaveTextContent("BTC-USD");
  });
});
