import React from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ExchangeHeader } from "./exchange-header";

describe("ExchangeHeader", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("calls pair search when opening the search modal", async () => {
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

    fireEvent.click(screen.getByTestId("pairs-open-button"));
    await vi.advanceTimersByTimeAsync(260);

    expect(onPairSearch).toHaveBeenCalledWith("");
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
  });
});
