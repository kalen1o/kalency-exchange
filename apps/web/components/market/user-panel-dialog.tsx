"use client";

import React from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export type UserPanelTab = "order" | "trades" | "orders";

export type UserPanelDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tab: UserPanelTab;
  onTabChange: (tab: UserPanelTab) => void;

  symbol: string;
  userId: string;
  onUserIdChange: (value: string) => void;

  quoteCurrency: string;

  rangePresets: readonly string[];
  rangePreset: string;
  onRangePresetChange: (value: string) => void;

  side: "BUY" | "SELL";
  type: "MARKET" | "LIMIT";
  qty: number;
  price: number;
  onSideChange: (value: "BUY" | "SELL") => void;
  onTypeChange: (value: "MARKET" | "LIMIT") => void;
  onQtyChange: (value: number) => void;
  onPriceChange: (value: number) => void;
  busy: boolean;
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;

  trades: Array<{ tradeId: string; price: number; qty: number }>;
  tradeSummary: { lastPrice: string; totalQty: number };

  orders: Array<{ orderId: string; side: string; symbol: string; remainingQty: number; qty: number }>;
  cancelOrderID: string | null;
  onCancelOrder: (orderId: string) => void;
};

export function UserPanelDialog({
  open,
  onOpenChange,
  tab,
  onTabChange,
  symbol,
  userId,
  onUserIdChange,
  quoteCurrency,
  rangePresets,
  rangePreset,
  onRangePresetChange,
  side,
  type,
  qty,
  price,
  onSideChange,
  onTypeChange,
  onQtyChange,
  onPriceChange,
  busy,
  onSubmit,
  trades,
  tradeSummary,
  orders,
  cancelOrderID,
  onCancelOrder
}: UserPanelDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-testid="user-panel">
        <DialogHeader>
          <DialogTitle className="font-mono">User Panel</DialogTitle>
          <DialogDescription>
            Manage orders and view activity for <span className="font-mono">{symbol.toUpperCase()}</span> as{" "}
            <span className="font-mono">{userId}</span>.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-wrap gap-2">
          <Button type="button" size="sm" variant={tab === "order" ? "default" : "outline"} onClick={() => onTabChange("order")} className="font-mono">
            Order
          </Button>
          <Button type="button" size="sm" variant={tab === "trades" ? "default" : "outline"} onClick={() => onTabChange("trades")} className="font-mono">
            Trades
          </Button>
          <Button type="button" size="sm" variant={tab === "orders" ? "default" : "outline"} onClick={() => onTabChange("orders")} className="font-mono">
            Orders
          </Button>
        </div>

        {tab === "order" && (
          <div className="grid gap-4">
            <div className="grid gap-3 sm:grid-cols-2">
              <div className="space-y-1.5">
                <Label htmlFor="user-id">User</Label>
                <Input id="user-id" value={userId} onChange={(e) => onUserIdChange(e.target.value)} className="font-mono" />
              </div>
              <div className="space-y-1.5">
                <Label>Pair</Label>
                <div className="rounded-md border border-border/70 bg-background/50 px-3 py-2 font-mono text-sm">{symbol.toUpperCase()}</div>
              </div>
            </div>

            <div className="flex flex-wrap gap-2">
              {rangePresets.map((preset) => (
                <Button
                  key={preset}
                  type="button"
                  size="sm"
                  variant={rangePreset === preset ? "default" : "outline"}
                  onClick={() => onRangePresetChange(preset)}
                >
                  {preset}
                </Button>
              ))}
            </div>

            <form onSubmit={onSubmit} className="grid gap-3">
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <Label htmlFor="order-side">Side</Label>
                  <Select value={side} onValueChange={(value) => onSideChange(value as "BUY" | "SELL")}>
                    <SelectTrigger id="order-side">
                      <SelectValue placeholder="Select side" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="BUY">BUY</SelectItem>
                      <SelectItem value="SELL">SELL</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="order-type">Type</Label>
                  <Select value={type} onValueChange={(value) => onTypeChange(value as "MARKET" | "LIMIT")}>
                    <SelectTrigger id="order-type">
                      <SelectValue placeholder="Select order type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="MARKET">MARKET</SelectItem>
                      <SelectItem value="LIMIT">LIMIT</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <Label htmlFor="order-qty">Quantity</Label>
                  <Input
                    id="order-qty"
                    type="number"
                    min={1}
                    value={qty}
                    onChange={(e) => onQtyChange(Number(e.target.value))}
                    required
                    className="font-mono"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="order-price">Price ({quoteCurrency})</Label>
                  <Input
                    id="order-price"
                    type="number"
                    min={1}
                    value={price}
                    onChange={(e) => onPriceChange(Number(e.target.value))}
                    disabled={type === "MARKET"}
                    required={type === "LIMIT"}
                    className="font-mono"
                  />
                </div>
              </div>

              <Button type="submit" disabled={busy} className="w-full">
                {busy ? "Submitting..." : "Submit Order"}
              </Button>
            </form>
          </div>
        )}

        {tab === "trades" && (
          <div className="space-y-3">
            <div className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-border/70 bg-background/40 px-3 py-2 font-mono text-xs">
              <span>
                Last: <span className="text-foreground">{tradeSummary.lastPrice}</span>
              </span>
              <span>
                Volume: <span className="text-foreground">{tradeSummary.totalQty}</span>
              </span>
            </div>
            <div className="space-y-2">
              {trades.length === 0 && <p className="text-sm text-muted-foreground">No trades yet.</p>}
              {trades.map((trade) => (
                <div key={trade.tradeId} className="grid grid-cols-[1fr_auto_auto] items-center rounded-md border border-border/70 bg-background/40 p-2 font-mono text-xs">
                  <span className="truncate pr-2">{trade.tradeId}</span>
                  <span className="px-2">{trade.price}</span>
                  <span>{trade.qty}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {tab === "orders" && (
          <div className="space-y-2">
            {orders.length === 0 && <p className="text-sm text-muted-foreground">No open orders.</p>}
            {orders.map((order) => (
              <div
                key={order.orderId}
                className="grid grid-cols-[auto_auto_minmax(0,1fr)_auto] items-center gap-2 rounded-md border border-border/70 bg-background/40 p-2 font-mono text-xs"
              >
                <span>{order.side}</span>
                <span>{order.symbol}</span>
                <span className="truncate">
                  {order.remainingQty}/{order.qty}
                </span>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  disabled={cancelOrderID === order.orderId}
                  onClick={() => onCancelOrder(order.orderId)}
                >
                  {cancelOrderID === order.orderId ? "Canceling..." : "Cancel"}
                </Button>
              </div>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

