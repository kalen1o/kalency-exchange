"use client";

import { useEffect } from "react";

export type UseMarketDocumentTitleParams = {
  symbol: string;
  price: number | null | undefined;
  changePct: number | null;
  suffix?: string;
};

export function useMarketDocumentTitle({
  symbol,
  price,
  changePct,
  suffix = "Kalency",
}: UseMarketDocumentTitleParams) {
  useEffect(() => {
    const symbolLabel = symbol.toUpperCase();
    const locale =
      typeof navigator === "undefined"
        ? "en-US"
        : (navigator.languages?.[0] || navigator.language || "en-US").trim() || "en-US";
    const priceText =
      price === null || price === undefined || !Number.isFinite(price)
        ? "-"
        : new Intl.NumberFormat(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(price);
    const pct =
      changePct === null || Number.isNaN(changePct)
        ? "-"
        : `${changePct > 0 ? "+" : changePct < 0 ? "-" : ""}${Math.abs(changePct).toFixed(2)}%`;
    document.title = `${symbolLabel} ${priceText} ${pct} | ${suffix}`;
  }, [symbol, price, changePct, suffix]);
}
