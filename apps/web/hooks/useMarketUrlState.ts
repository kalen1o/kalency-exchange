"use client";

import { useCallback } from "react";
import { parseAsFloat, parseAsStringEnum, useQueryStates } from "nuqs";
import type { ChartInterval } from "@/lib/api";

export type MarketChartViewState = {
  barSpacing: number;
  scrollPosition: number;
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
      p: parseAsFloat
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
      void setState({
        z: view.barSpacing ?? null,
        p: view.scrollPosition ?? null
      });
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
      scrollPosition: state.p ?? defaultView.scrollPosition ?? 0
    } satisfies MarketChartViewState,
    initialView: {
      barSpacing: state.z ?? undefined,
      scrollPosition: state.p ?? undefined
    },
    setView
  };
}
