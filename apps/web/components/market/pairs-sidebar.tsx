"use client";

import React from "react";
import { Button } from "@/components/ui/button";

export type PairsSidebarProps = {
  pairs: readonly string[];
  selected: string;
  onSelect: (pair: string) => void;
};

export function PairsSidebar({ pairs, selected, onSelect }: PairsSidebarProps) {
  return (
    <aside data-testid="pairs-sidebar" className="h-full min-h-0 overflow-y-auto p-3">
      <div className="flex items-center justify-between px-1 py-1.5">
        <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Pairs</p>
        <p className="font-mono text-[11px] text-muted-foreground">{pairs.length}</p>
      </div>
      <div className="grid gap-2">
        {pairs.map((pair) => (
          <Button
            key={pair}
            type="button"
            variant={selected === pair ? "default" : "outline"}
            className="justify-between font-mono"
            onClick={() => onSelect(pair)}
          >
            <span>{pair}</span>
            <span className="text-xs text-muted-foreground">{pair.split("-")[1] ?? "USD"}</span>
          </Button>
        ))}
      </div>
    </aside>
  );
}
