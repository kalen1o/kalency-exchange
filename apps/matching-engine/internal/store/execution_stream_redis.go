package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/matching-engine/internal/matching"
)

type RedisExecutionStreamSink struct {
	client redis.UniversalClient
	stream string
}

func NewRedisExecutionStreamSink(client redis.UniversalClient, stream string) *RedisExecutionStreamSink {
	if stream == "" {
		stream = "kalency:v1:stream:executions"
	}
	return &RedisExecutionStreamSink{client: client, stream: stream}
}

func (s *RedisExecutionStreamSink) PublishExecution(ctx context.Context, execution matching.Execution) error {
	values := map[string]any{
		"trade_id":       execution.TradeID,
		"symbol":         execution.Symbol,
		"price":          execution.Price,
		"qty":            execution.Qty,
		"maker_order_id": execution.MakerOrderID,
		"maker_user_id":  execution.MakerUserID,
		"taker_order_id": execution.TakerOrderID,
		"taker_user_id":  execution.TakerUserID,
		"ts":             execution.TS.Format(time.RFC3339Nano),
	}

	return s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: s.stream,
		ID:     "*",
		Values: values,
	}).Err()
}
