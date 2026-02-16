export type DebouncedFn<Args extends unknown[]> = ((...args: Args) => void) & {
  cancel: () => void;
  flush: () => void;
};

export function debounce<Args extends unknown[]>(
  fn: (...args: Args) => void,
  waitMs: number
): DebouncedFn<Args> {
  let timer: ReturnType<typeof setTimeout> | null = null;
  let lastArgs: Args | null = null;

  const cancel = () => {
    if (timer !== null) {
      clearTimeout(timer);
      timer = null;
    }
    lastArgs = null;
  };

  const flush = () => {
    if (timer !== null) {
      clearTimeout(timer);
      timer = null;
    }
    if (!lastArgs) return;
    const args = lastArgs;
    lastArgs = null;
    fn(...args);
  };

  const debounced = (...args: Args) => {
    lastArgs = args;
    if (timer !== null) {
      clearTimeout(timer);
    }
    timer = setTimeout(() => {
      timer = null;
      flush();
    }, waitMs);
  };

  debounced.cancel = cancel;
  debounced.flush = flush;
  return debounced;
}

