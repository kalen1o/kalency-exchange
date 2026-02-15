CREATE TABLE IF NOT EXISTS trade_ledger (
  trade_id TEXT PRIMARY KEY,
  symbol TEXT NOT NULL,
  buy_user_id TEXT,
  sell_user_id TEXT,
  price DOUBLE PRECISION NOT NULL,
  qty DOUBLE PRECISION NOT NULL,
  executed_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_trade_ledger_symbol_executed_at
  ON trade_ledger(symbol, executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_trade_ledger_buy_user_executed_at
  ON trade_ledger(buy_user_id, executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_trade_ledger_sell_user_executed_at
  ON trade_ledger(sell_user_id, executed_at DESC);
