export type Side = "BUY" | "SELL";
export type OrderType = "MARKET" | "LIMIT";

export type DraftOrderInput = {
  clientOrderId: string;
  userId: string;
  symbol: string;
  side: Side;
  type: OrderType;
  qty: number;
  price?: number;
};

export type PlaceOrderPayload = {
  clientOrderId: string;
  userId: string;
  symbol: string;
  side: Side;
  type: OrderType;
  qty: number;
  price?: number;
};

export type OrderAck = {
  orderId: string;
  status: string;
  filledQty: number;
  remainingQty: number;
  avgPrice: number;
  clientOrderId?: string;
  symbol?: string;
  ts: string;
};

export type OpenOrder = {
  orderId: string;
  clientOrderId: string;
  userId: string;
  symbol: string;
  side: Side;
  type: OrderType;
  price: number;
  qty: number;
  remainingQty: number;
  createdAt: string;
};

export type Trade = {
  tradeId: string;
  symbol: string;
  price: number;
  qty: number;
  makerOrderId?: string;
  makerUserId?: string;
  makerUserID?: string;
  takerOrderId?: string;
  takerUserId?: string;
  takerUserID?: string;
  ts: string;
};

export type Tick = {
  symbol: string;
  price: number;
  volume?: number;
  ts: string;
};

export type Candle = {
  symbol: string;
  timeframe: string;
  bucketStart: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
};

export type ChartInterval =
  | "1m"
  | "3m"
  | "5m"
  | "15m"
  | "30m"
  | "1h"
  | "2h"
  | "4h"
  | "6h"
  | "8h"
  | "12h"
  | "1d"
  | "3d"
  | "1w"
  | "1M";

export const BINANCE_CHART_INTERVALS: ChartInterval[] = [
  "1m",
  "5m",
  "1h",
];

type BackendTimeframe = "1s" | "5s" | "1m" | "5m" | "1h";

const intervalToBackendTimeframe: Record<ChartInterval, BackendTimeframe> = {
  "1m": "1m",
  "3m": "1m",
  "5m": "5m",
  "15m": "5m",
  "30m": "5m",
  "1h": "1h",
  "2h": "1h",
  "4h": "1h",
  "6h": "1h",
  "8h": "1h",
  "12h": "1h",
  "1d": "1h",
  "3d": "1h",
  "1w": "1h",
  "1M": "1h"
};

export function mapChartIntervalToBackendTimeframe(interval: ChartInterval): BackendTimeframe {
  return intervalToBackendTimeframe[interval];
}

export type ChartRangePreset = "15m" | "1h" | "4h" | "24h";

type BinanceExchangeSymbol = {
  symbol: string;
  status: string;
  isSpotTradingAllowed?: boolean;
};

const FIAT_EQUIVALENT_QUOTES = new Set(["USDT", "USDC", "FDUSD", "BUSD", "USD"]);
const KNOWN_QUOTES = ["USDT", "USDC", "FDUSD", "BUSD", "USD", "BTC", "ETH", "BNB", "EUR", "TRY"] as const;

export function normalizeSymbol(input: string): string {
  return input.trim().toUpperCase();
}

function authHeaders(): Record<string, string> {
  const headers: Record<string, string> = {};
  const apiKey = (process.env.NEXT_PUBLIC_API_KEY ?? "").trim();
  const bearerToken = (process.env.NEXT_PUBLIC_BEARER_TOKEN ?? "").trim();
  if (apiKey !== "") {
    headers["X-API-Key"] = apiKey;
  }
  if (bearerToken !== "") {
    headers.Authorization = `Bearer ${bearerToken}`;
  }
  return headers;
}

export function buildOrderPayload(input: DraftOrderInput): PlaceOrderPayload {
  const payload: PlaceOrderPayload = {
    clientOrderId: input.clientOrderId.trim(),
    userId: input.userId.trim(),
    symbol: normalizeSymbol(input.symbol),
    side: input.side,
    type: input.type,
    qty: Number(input.qty),
  };

  if (input.type === "LIMIT") {
    payload.price = Number(input.price ?? 0);
  }

  return payload;
}

export function summarizeTrades(trades: Trade[]): { lastPrice: number | null; totalQty: number } {
  if (trades.length === 0) {
    return { lastPrice: null, totalQty: 0 };
  }

  return {
    lastPrice: trades[trades.length - 1].price,
    totalQty: trades.reduce((acc, trade) => acc + trade.qty, 0),
  };
}

export function rangeFromPreset(preset: ChartRangePreset, now = new Date()): { from: string; to: string } {
  const to = new Date(now.getTime());
  const presetMs: Record<ChartRangePreset, number> = {
    "15m": 15 * 60 * 1000,
    "1h": 60 * 60 * 1000,
    "4h": 4 * 60 * 60 * 1000,
    "24h": 24 * 60 * 60 * 1000,
  };
  const from = new Date(to.getTime() - presetMs[preset]);
  return { from: from.toISOString(), to: to.toISOString() };
}

function normalizeApiBase(base: string): string {
  const trimmed = base.trim();
  if (!trimmed) {
    return "http://localhost:8080";
  }
  return trimmed.endsWith("/") ? trimmed.slice(0, -1) : trimmed;
}

async function parseJSON<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `request failed with ${res.status}`);
  }
  return (await res.json()) as T;
}

export async function placeOrder(base: string, input: DraftOrderInput): Promise<OrderAck> {
  const apiBase = normalizeApiBase(base);
  const payload = buildOrderPayload(input);

  const res = await fetch(`${apiBase}/v1/orders`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(payload),
  });

  return parseJSON<OrderAck>(res);
}

export async function cancelOrder(base: string, orderId: string): Promise<OrderAck> {
  const apiBase = normalizeApiBase(base);
  const safeOrderID = encodeURIComponent(orderId.trim());
  const res = await fetch(`${apiBase}/v1/orders/${safeOrderID}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
  return parseJSON<OrderAck>(res);
}

export async function fetchOpenOrders(base: string): Promise<OpenOrder[]> {
  const apiBase = normalizeApiBase(base);
  const res = await fetch(`${apiBase}/v1/orders/open`, { headers: authHeaders() });
  return parseJSON<OpenOrder[]>(res);
}

export async function ensureSimulatedSymbol(base: string, symbol: string): Promise<void> {
  const apiBase = normalizeApiBase(base);
  const safeSymbol = encodeURIComponent(normalizeSymbol(symbol));
  const res = await fetch(`${apiBase}/v1/admin/symbols/${safeSymbol}/ensure`, {
    method: "POST",
    headers: authHeaders()
  });
  await parseJSON<Record<string, unknown>>(res);
}

export async function fetchTrades(base: string, symbol: string, limit = 50): Promise<Trade[]> {
  const apiBase = normalizeApiBase(base);
  const safeSymbol = encodeURIComponent(normalizeSymbol(symbol));
  const res = await fetch(`${apiBase}/v1/markets/${safeSymbol}/trades?limit=${limit}`, {
    headers: authHeaders(),
  });
  return parseJSON<Trade[]>(res);
}

export function buildTradesWebSocketURL(base: string, symbol: string): string {
  const apiBase = normalizeApiBase(base);
  const url = new URL(apiBase);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = `/ws/trades/${encodeURIComponent(normalizeSymbol(symbol))}`;
  url.search = "";
  url.hash = "";
  return url.toString();
}

export function buildTicksWebSocketURL(base: string, symbol: string): string {
  const apiBase = normalizeApiBase(base);
  const url = new URL(apiBase);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = `/ws/ticks/${encodeURIComponent(normalizeSymbol(symbol))}`;
  url.search = "";
  url.hash = "";
  return url.toString();
}

export async function fetchCandles(
  base: string,
  symbol: string,
  timeframe: string,
  from?: string,
  to?: string
): Promise<Candle[]> {
  const apiBase = normalizeApiBase(base);
  const safeSymbol = encodeURIComponent(normalizeSymbol(symbol));
  const query = new URLSearchParams();
  query.set("tf", timeframe.trim().toLowerCase());
  if (from?.trim()) {
    query.set("from", from.trim());
  }
  if (to?.trim()) {
    query.set("to", to.trim());
  }

  const res = await fetch(`${apiBase}/v1/markets/${safeSymbol}/candles?${query.toString()}`, {
    headers: authHeaders(),
  });
  return parseJSON<Candle[]>(res);
}

export async function searchPairsByAPI(query: string, fallbackPairs: readonly string[] = [], limit = 20): Promise<string[]> {
  const normalizedQuery = query.trim().toUpperCase().replace(/[^A-Z0-9]/g, "");
  const matchesQuery = (pair: string) =>
    normalizedQuery.length === 0 || pair.replace(/[^A-Z0-9]/g, "").includes(normalizedQuery);
  const fallbackMatches = fallbackPairs
    .map((pair) => normalizeSymbol(pair))
    .filter((pair) => matchesQuery(pair))
    .slice(0, limit);

  try {
    const res = await fetch("https://api.binance.com/api/v3/exchangeInfo");
    const body = await parseJSON<{ symbols?: BinanceExchangeSymbol[] }>(res);
    const remotePairs = (body.symbols ?? [])
      .filter((item) => item.status === "TRADING" && (item.isSpotTradingAllowed ?? true))
      .map((item) => {
        for (const quote of KNOWN_QUOTES) {
          if (item.symbol.endsWith(quote) && item.symbol.length > quote.length) {
            const base = item.symbol.slice(0, item.symbol.length - quote.length);
            const normalizedQuote = FIAT_EQUIVALENT_QUOTES.has(quote) ? "USD" : quote;
            return `${base}-${normalizedQuote}`;
          }
        }
        return null;
      })
      .filter((pair): pair is string => pair !== null)
      .filter((pair) => matchesQuery(pair));

    const merged = Array.from(new Set([...fallbackMatches, ...remotePairs]));
    return merged.slice(0, limit);
  } catch {
    return fallbackMatches;
  }
}
