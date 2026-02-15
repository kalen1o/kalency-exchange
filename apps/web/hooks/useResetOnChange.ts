"use client";

import { useEffect } from "react";

export function useResetOnChange(reset: () => void, deps: readonly unknown[]) {
  useEffect(() => {
    reset();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);
}

