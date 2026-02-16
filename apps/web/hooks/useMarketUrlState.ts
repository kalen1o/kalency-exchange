"use client";

import { useCallback } from "react";
import { parseAsFloat, parseAsStringEnum, useQueryStates } from "nuqs";
import type { ChartInterval } from "@/lib/api";

export type MarketChartViewState = {
  barSpacing: number;
  scrollPosition: number;
  priceFrom: number | null;
  priceTo: number | null;
};

export type UseMarketUrlStateParams = {
  pairOptions: readonly string[];
  timeframeOptions: readonly ChartInterval[];
  defaultPair: string;
  defaultTimeframe: ChartInterval;
  defaultView?: Partial<MarketChartViewState>;
};

export function useMarketUrlState({
  pairOptions,
  timeframeOptions,
  defaultPair,
  defaultTimeframe,
  defaultView = { barSpacing: 10, scrollPosition: 0 }
}: UseMarketUrlStateParams) {
  const [state, setState] = useQueryStates(
    {
      pair: parseAsStringEnum([...pairOptions]).withDefault(defaultPair),
      tf: parseAsStringEnum([...timeframeOptions]).withDefault(defaultTimeframe),
      z: parseAsFloat,
      p: parseAsFloat,
      yf: parseAsFloat,
      yt: parseAsFloat
    },
    { history: "replace" }
  );

  const setPair = useCallback(
    (pair: string) => {
      void setState({ pair });
    },
    [setState]
  );

  const setTimeframe = useCallback(
    (tf: ChartInterval) => {
      void setState({ tf });
    },
    [setState]
  );

  const setView = useCallback(
    (view: Partial<MarketChartViewState>) => {
      const nextState: {
        z?: number | null;
        p?: number | null;
        yf?: number | null;
        yt?: number | null;
      } = {};

      if ("barSpacing" in view) {
        nextState.z = view.barSpacing ?? null;
      }
      if ("scrollPosition" in view) {
        nextState.p = view.scrollPosition ?? null;
      }
      if ("priceFrom" in view) {
        nextState.yf = view.priceFrom ?? null;
      }
      if ("priceTo" in view) {
        nextState.yt = view.priceTo ?? null;
      }

      void setState(nextState);
    },
    [setState]
  );

  return {
    pair: state.pair,
    setPair,
    timeframe: state.tf as ChartInterval,
    setTimeframe,
    view: {
      barSpacing: state.z ?? defaultView.barSpacing ?? 10,
      scrollPosition: state.p ?? defaultView.scrollPosition ?? 0,
      priceFrom: state.yf ?? null,
      priceTo: state.yt ?? null
    } satisfies MarketChartViewState,
    initialView: {
      barSpacing: state.z ?? undefined,
      scrollPosition: state.p ?? undefined,
      priceFrom: state.yf ?? undefined,
      priceTo: state.yt ?? undefined
    },
    setView
  };
}
