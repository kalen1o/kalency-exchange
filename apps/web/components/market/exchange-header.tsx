"use client";

import React, { useEffect, useMemo, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export type HeaderPanelTab = "order" | "trades" | "orders";

export type ExchangeHeaderProps = {
  symbol: string;
  timeframeOptions: readonly string[];
  timeframe: string;
  onTimeframeChange: (value: string) => void;

  userId: string;
  onOpenPanel: (tab: HeaderPanelTab) => void;
};

export function ExchangeHeader({
  symbol,
  timeframeOptions,
  timeframe,
  onTimeframeChange,
  userId,
  onOpenPanel
}: ExchangeHeaderProps) {
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement | null>(null);

  const avatarLabel = useMemo(() => {
    const trimmed = userId.trim();
    if (trimmed === "") return "U";
    const first = trimmed[0] ?? "U";
    return String(first).toUpperCase();
  }, [userId]);

  useEffect(() => {
    if (!userMenuOpen) return;

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setUserMenuOpen(false);
      }
    };

    const onPointerDown = (event: PointerEvent) => {
      const root = userMenuRef.current;
      if (!root) return;
      if (event.target instanceof Node && root.contains(event.target)) return;
      setUserMenuOpen(false);
    };

    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("pointerdown", onPointerDown);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("pointerdown", onPointerDown);
    };
  }, [userMenuOpen]);

  return (
    <div
      data-testid="top-header"
      className="flex items-stretch justify-between border-b border-border/70 bg-card/70"
    >
      <div className="flex min-w-0 flex-1 items-stretch divide-x divide-border/70 overflow-x-auto">
        <div className="flex shrink-0 items-center gap-2 px-3 py-2">
          <p className="font-mono text-sm font-semibold tracking-wide">{symbol.toUpperCase()}</p>
        </div>

        <div className="flex shrink-0 items-center px-3 py-2">
          <Select value={timeframe} onValueChange={onTimeframeChange}>
            <SelectTrigger className="h-8 w-[72px] rounded-none border-0 bg-transparent px-0 py-0 font-mono text-xs shadow-none focus:ring-0 focus:ring-offset-0">
              <SelectValue placeholder="TF" />
            </SelectTrigger>
            <SelectContent>
              {timeframeOptions.map((interval) => (
                <SelectItem key={interval} value={interval} className="font-mono text-xs">
                  {interval}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div ref={userMenuRef} className="relative flex shrink-0 items-center border-l border-border/70 px-2">
        <Button
          data-testid="user-avatar"
          type="button"
          variant="ghost"
          size="sm"
          className="h-10 w-10 rounded-full p-0 font-mono text-sm"
          onClick={() => setUserMenuOpen((open) => !open)}
        >
          {avatarLabel}
        </Button>

        {userMenuOpen && (
          <div
            data-testid="user-menu"
            className="absolute right-0 top-12 z-50 w-56 overflow-hidden rounded-md border border-border/70 bg-card shadow-lg"
          >
            <button
              type="button"
              className="flex w-full items-center justify-between px-3 py-2 text-left text-sm hover:bg-background/60"
              onClick={() => {
                setUserMenuOpen(false);
                onOpenPanel("order");
              }}
            >
              <span className="font-mono">Order Ticket</span>
              <span className="text-xs text-muted-foreground">Enter</span>
            </button>
            <button
              type="button"
              className="flex w-full items-center justify-between px-3 py-2 text-left text-sm hover:bg-background/60"
              onClick={() => {
                setUserMenuOpen(false);
                onOpenPanel("trades");
              }}
            >
              <span className="font-mono">Recent Trades</span>
              <span className="text-xs text-muted-foreground">Enter</span>
            </button>
            <button
              type="button"
              className="flex w-full items-center justify-between px-3 py-2 text-left text-sm hover:bg-background/60"
              onClick={() => {
                setUserMenuOpen(false);
                onOpenPanel("orders");
              }}
            >
              <span className="font-mono">Open Orders</span>
              <span className="text-xs text-muted-foreground">Enter</span>
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
