import React from "react";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ExchangeHeader } from "./exchange-header";

describe("ExchangeHeader", () => {
  afterEach(() => {
    cleanup();
    vi.useRealTimers();
  });

  it("does not call pair search when opening the search modal", async () => {
    vi.useFakeTimers();
    const onPairSearch = vi.fn().mockResolvedValue(["BTC-USD", "ETH-USD"]);

    render(
      <ExchangeHeader
        symbol="BTC-USD"
        timeframeOptions={["1m", "5m"]}
        timeframe="1m"
        onTimeframeChange={vi.fn()}
        pairOptions={["BTC-USD", "ETH-USD"]}
        onPairSelect={vi.fn()}
        onPairSearch={onPairSearch}
        barStyleOptions={["candles", "bars", "line", "line-area"]}
        barStyle="candles"
        onBarStyleChange={vi.fn()}
        showVolume
        onToggleVolume={vi.fn()}
        userId="demo-user"
        onOpenPanel={vi.fn()}
      />
    );

    fireEvent.click(screen.getAllByTestId("pairs-open-button")[0]);
    await vi.advanceTimersByTimeAsync(260);

    expect(onPairSearch).not.toHaveBeenCalled();
  });

  it("calls pair search when user types in the search modal", async () => {
    vi.useFakeTimers();
    const onPairSearch = vi.fn().mockResolvedValue(["BTC-USD", "ETH-USD"]);

    render(
      <ExchangeHeader
        symbol="BTC-USD"
        timeframeOptions={["1m", "5m"]}
        timeframe="1m"
        onTimeframeChange={vi.fn()}
        pairOptions={["BTC-USD", "ETH-USD"]}
        onPairSelect={vi.fn()}
        onPairSearch={onPairSearch}
        barStyleOptions={["candles", "bars", "line", "line-area"]}
        barStyle="candles"
        onBarStyleChange={vi.fn()}
        showVolume
        onToggleVolume={vi.fn()}
        userId="demo-user"
        onOpenPanel={vi.fn()}
      />
    );

    fireEvent.click(screen.getAllByTestId("pairs-open-button")[0]);
    fireEvent.change(screen.getByTestId("pairs-search-input"), { target: { value: "btc" } });
    await vi.advanceTimersByTimeAsync(260);

    expect(onPairSearch).toHaveBeenCalledWith("btc");
  });

  it("opens trades panel from header trades button", () => {
    const onOpenPanel = vi.fn();

    render(
      <ExchangeHeader
        symbol="BTC-USD"
        timeframeOptions={["1m", "5m"]}
        timeframe="1m"
        onTimeframeChange={vi.fn()}
        pairOptions={["BTC-USD", "ETH-USD"]}
        onPairSelect={vi.fn()}
        onPairSearch={vi.fn().mockResolvedValue(["BTC-USD", "ETH-USD"])}
        barStyleOptions={["candles", "bars", "line", "line-area"]}
        barStyle="candles"
        onBarStyleChange={vi.fn()}
        showVolume
        onToggleVolume={vi.fn()}
        userId="demo-user"
        onOpenPanel={onOpenPanel}
      />
    );

    fireEvent.click(screen.getByTestId("header-trades-button"));
    expect(onOpenPanel).toHaveBeenCalledWith("trades");
  });

  it("shows chart style icons for each selected style", () => {
    const styles = ["candles", "bars", "line", "line-area"] as const;

    for (const style of styles) {
      const view = render(
        <ExchangeHeader
          symbol="BTC-USD"
          timeframeOptions={["1m", "5m"]}
          timeframe="1m"
          onTimeframeChange={vi.fn()}
          pairOptions={["BTC-USD", "ETH-USD"]}
          onPairSelect={vi.fn()}
          onPairSearch={vi.fn().mockResolvedValue(["BTC-USD", "ETH-USD"])}
          barStyleOptions={styles}
          barStyle={style}
          onBarStyleChange={vi.fn()}
          showVolume
          onToggleVolume={vi.fn()}
          userId="demo-user"
          onOpenPanel={vi.fn()}
        />
      );

      expect(screen.getAllByTestId(`chart-style-icon-${style}-trigger`).length).toBeGreaterThan(0);
      view.unmount();
    }
  });

  it("shows a small logo in the header", () => {
    render(
      <ExchangeHeader
        symbol="BTC-USD"
        timeframeOptions={["1m", "5m"]}
        timeframe="1m"
        onTimeframeChange={vi.fn()}
        pairOptions={["BTC-USD", "ETH-USD"]}
        onPairSelect={vi.fn()}
        onPairSearch={vi.fn().mockResolvedValue(["BTC-USD", "ETH-USD"])}
        barStyleOptions={["candles", "bars", "line", "line-area"]}
        barStyle="candles"
        onBarStyleChange={vi.fn()}
        showVolume
        onToggleVolume={vi.fn()}
        userId="demo-user"
        onOpenPanel={vi.fn()}
      />
    );

    const logos = screen.getAllByTestId("header-small-logo");
    expect(logos.length).toBeGreaterThan(0);
    expect(logos[0]).toHaveAttribute("src", "/small-logo.png");
    expect(logos[0]).toHaveClass("rounded-full");
    expect(logos[0]).toHaveClass("h-8", "w-8", "object-cover");
  });
});
