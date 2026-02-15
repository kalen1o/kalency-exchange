package store

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"kalency/apps/ledger-writer/internal/ledger"
)

type PostgresSink struct {
	pool *pgxpool.Pool
}

func NewPostgresSink(ctx context.Context, dsn string) (*PostgresSink, error) {
	pool, err := pgxpool.New(ctx, strings.TrimSpace(dsn))
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresSink{pool: pool}, nil
}

func (s *PostgresSink) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *PostgresSink) WriteExecution(ctx context.Context, event ledger.ExecutionEvent) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO trade_ledger (
			trade_id,
			symbol,
			buy_user_id,
			sell_user_id,
			price,
			qty,
			executed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (trade_id) DO NOTHING
	`,
		event.TradeID,
		event.Symbol,
		event.BuyUserID,
		event.SellUserID,
		event.Price,
		event.Qty,
		event.ExecutedAt,
	)
	return err
}

type LogSink struct{}

func (LogSink) WriteExecution(context.Context, ledger.ExecutionEvent) error {
	return nil
}
