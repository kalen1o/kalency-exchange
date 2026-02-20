import React from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { UserPanelDialog, type UserPanelDialogProps } from "./user-panel-dialog";

function buildProps(overrides: Partial<UserPanelDialogProps> = {}): UserPanelDialogProps {
  const trades = Array.from({ length: 40 }, (_, i) => ({
    tradeId: `trade-${i + 1}`,
    price: 100 + i,
    qty: i + 1,
    inUser: `in-${i + 1}`,
    outUser: `out-${i + 1}`
  }));

  return {
    open: true,
    onOpenChange: vi.fn(),
    tab: "trades",
    onTabChange: vi.fn(),
    symbol: "BTC-USD",
    userId: "demo-user",
    onUserIdChange: vi.fn(),
    quoteCurrency: "USD",
    rangePresets: ["15m", "1h", "4h", "24h"],
    rangePreset: "1h",
    onRangePresetChange: vi.fn(),
    side: "BUY",
    type: "MARKET",
    qty: 1,
    price: 100,
    onSideChange: vi.fn(),
    onTypeChange: vi.fn(),
    onQtyChange: vi.fn(),
    onPriceChange: vi.fn(),
    busy: false,
    onSubmit: vi.fn(),
    trades,
    tradeSummary: { lastPrice: "100.00 USD", totalQty: 100 },
    orders: [],
    cancelOrderID: null,
    onCancelOrder: vi.fn(),
    ...overrides
  };
}

describe("UserPanelDialog", () => {
  it("loads more trades when scrolled near bottom", async () => {
    render(<UserPanelDialog {...buildProps()} />);

    expect(screen.getByText("in-20")).toBeInTheDocument();
    expect(screen.queryByText("in-21")).not.toBeInTheDocument();

    const scroller = screen.getByTestId("trades-scroll-container");
    Object.defineProperty(scroller, "clientHeight", { value: 200, configurable: true });
    Object.defineProperty(scroller, "scrollHeight", { value: 1000, configurable: true });
    Object.defineProperty(scroller, "scrollTop", { value: 800, configurable: true });

    fireEvent.scroll(scroller);

    await waitFor(() => {
      expect(screen.getByText("in-21")).toBeInTheDocument();
    });
  });
});
