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
});
