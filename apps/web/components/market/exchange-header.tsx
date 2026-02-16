"use client";

import React, { useEffect, useMemo, useRef, useState } from "react";
import { BarChart3, SquareFunction, TrendingUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { ChartBarStyle } from "@/components/market/tradingview-chart";

export type HeaderPanelTab = "order" | "trades" | "orders";

export type ExchangeHeaderProps = {
  symbol: string;
  timeframeOptions: readonly string[];
  timeframe: string;
  onTimeframeChange: (value: string) => void;
  pairOptions: readonly string[];
  onPairSelect: (value: string) => void;
  onPairSearch: (query: string) => Promise<string[]>;
  barStyleOptions: readonly ChartBarStyle[];
  barStyle: ChartBarStyle;
  onBarStyleChange: (value: ChartBarStyle) => void;
  showVolume: boolean;
  onToggleVolume: (next: boolean) => void;

  userId: string;
  onOpenPanel: (tab: HeaderPanelTab) => void;
};

type BarStyleIconProps = {
  style: ChartBarStyle;
  testId: string;
  className?: string;
};

function BarStyleIcon({ style, testId, className = "h-3.5 w-3.5" }: BarStyleIconProps) {
  if (style === "candles") {
    return (
      <svg
        data-testid={testId}
        viewBox="0 0 24 24"
        className={className}
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden
      >
        <path d="M7 4v16" />
        <path d="M17 4v16" />
        <rect x="5.5" y="8" width="3" height="6" rx="0.6" fill="currentColor" stroke="none" />
        <rect x="15.5" y="10" width="3" height="4" rx="0.6" fill="currentColor" stroke="none" />
      </svg>
    );
  }
  if (style === "bars") {
    return <BarChart3 data-testid={testId} className={className} aria-hidden />;
  }
  if (style === "line") {
    return <TrendingUp data-testid={testId} className={className} aria-hidden />;
  }
  return (
    <svg
      data-testid={testId}
      viewBox="0 0 24 24"
      className={className}
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M4 16L9 12L13 14L20 8" />
      <path d="M4 20V16L9 12L13 14L20 8V20Z" fill="currentColor" opacity="0.2" stroke="none" />
    </svg>
  );
}

export function ExchangeHeader({
  symbol,
  timeframeOptions,
  timeframe,
  onTimeframeChange,
  pairOptions,
  onPairSelect,
  onPairSearch,
  barStyleOptions,
  barStyle,
  onBarStyleChange,
  showVolume,
  onToggleVolume,
  userId,
  onOpenPanel
}: ExchangeHeaderProps) {
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [indicatorsOpen, setIndicatorsOpen] = useState(false);
  const [pairsOpen, setPairsOpen] = useState(false);
  const [pairQuery, setPairQuery] = useState("");
  const [pairResults, setPairResults] = useState<string[]>([...pairOptions]);
  const [pairLoading, setPairLoading] = useState(false);
  const onPairSearchRef = useRef(onPairSearch);
  const userMenuRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    onPairSearchRef.current = onPairSearch;
  }, [onPairSearch]);

  const avatarLabel = useMemo(() => {
    const trimmed = userId.trim();
    if (trimmed === "") return "U";
    const first = trimmed[0] ?? "U";
    return String(first).toUpperCase();
  }, [userId]);

  useEffect(() => {
    if (!pairsOpen) return;
    const trimmed = pairQuery.trim();
    let active = true;
    const timer = window.setTimeout(async () => {
      setPairLoading(true);
      try {
        const next = await onPairSearchRef.current(trimmed);
        if (!active) return;
        setPairResults(next.length === 0 && trimmed === "" ? [...pairOptions] : next);
      } finally {
        if (active) {
          setPairLoading(false);
        }
      }
    }, 250);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [pairsOpen, pairQuery, pairOptions]);

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
          <img
            data-testid="header-small-logo"
            src="/small-logo.png"
            alt="Kalency logo"
            className="h-5 w-5 rounded-sm object-contain"
          />
          <button
            type="button"
            data-testid="pairs-open-button"
            className="font-mono text-sm font-semibold tracking-wide hover:text-primary"
            onClick={() => {
              setPairQuery("");
              setPairResults([...pairOptions]);
              setPairsOpen(true);
            }}
          >
            {symbol.toUpperCase()}
          </button>
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
        <div className="flex shrink-0 items-center px-3 py-2">
          <Select value={barStyle} onValueChange={(value) => onBarStyleChange(value as ChartBarStyle)}>
            <SelectTrigger
              data-testid="chart-bar-style-select"
              className="h-8 w-[108px] rounded-none border-0 bg-transparent px-0 py-0 font-mono text-xs shadow-none focus:ring-0 focus:ring-offset-0"
            >
              <div className="flex items-center gap-1.5">
                <BarStyleIcon style={barStyle} testId={`chart-style-icon-${barStyle}-trigger`} />
                <span className="capitalize">{barStyle.replace("-", " ")}</span>
              </div>
            </SelectTrigger>
            <SelectContent>
              {barStyleOptions.map((style) => (
                <SelectItem key={style} value={style} className="font-mono text-xs capitalize">
                  <div className="flex items-center gap-2">
                    <BarStyleIcon style={style} testId={`chart-style-icon-${style}-option`} />
                    <span>{style.replace("-", " ")}</span>
                  </div>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="relative flex shrink-0 items-center px-3 py-2">
          <button
            type="button"
            data-testid="chart-indicators-button"
            aria-expanded={indicatorsOpen}
            aria-label="Indicators"
            onClick={() => setIndicatorsOpen((open) => !open)}
            className="inline-flex items-center rounded border border-border/60 bg-background/40 px-2 py-1 text-foreground hover:bg-background/60"
          >
            <SquareFunction className="h-3.5 w-3.5" aria-hidden />
          </button>
          <Dialog open={indicatorsOpen} onOpenChange={setIndicatorsOpen}>
            <DialogContent data-testid="chart-indicators-menu" className="w-[320px] max-w-[92vw] p-4">
              <DialogHeader>
                <DialogTitle className="font-mono text-sm">indicator</DialogTitle>
                <DialogDescription className="font-mono text-xs">volums</DialogDescription>
              </DialogHeader>
              <div className="flex items-center justify-between gap-3">
                <span className="font-mono text-xs text-foreground">volums</span>
                <button
                  type="button"
                  role="switch"
                  data-testid="chart-indicators-volume-switch"
                  aria-checked={showVolume}
                  onClick={() => onToggleVolume(!showVolume)}
                  className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
                    showVolume ? "bg-emerald-500/80" : "bg-white/20"
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 rounded-full bg-white transition-transform ${
                      showVolume ? "translate-x-4" : "translate-x-0.5"
                    }`}
                  />
                </button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <Dialog open={pairsOpen} onOpenChange={setPairsOpen}>
        <DialogContent data-testid="pairs-search-modal" className="w-[420px] max-w-[94vw] p-4">
          <DialogHeader>
            <DialogTitle className="font-mono text-sm">pairs</DialogTitle>
            <DialogDescription className="font-mono text-xs">Search pair from API</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <Input
              data-testid="pairs-search-input"
              value={pairQuery}
              onChange={(event) => setPairQuery(event.target.value)}
              placeholder="Search pair (ex: btc, eth)"
              className="h-9 font-mono text-xs"
            />
            <div className="max-h-64 space-y-1 overflow-auto rounded border border-border/70 p-1">
              {pairLoading ? (
                <p className="px-2 py-1 font-mono text-xs text-muted-foreground">Searching...</p>
              ) : pairResults.length === 0 ? (
                <p className="px-2 py-1 font-mono text-xs text-muted-foreground">No pairs found.</p>
              ) : (
                pairResults.map((pair) => (
                  <button
                    key={pair}
                    type="button"
                    className="flex w-full items-center justify-between rounded px-2 py-1.5 text-left font-mono text-xs hover:bg-background/60"
                    onClick={() => {
                      onPairSelect(pair);
                      setPairsOpen(false);
                    }}
                  >
                    <span>{pair}</span>
                    <span className="text-[10px] text-muted-foreground">{pair.split("-")[1] ?? "USD"}</span>
                  </button>
                ))
              )}
            </div>
          </div>
        </DialogContent>
      </Dialog>

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
