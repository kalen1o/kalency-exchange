package store

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/market-sim/internal/sim"
)

type RedisTickStreamSink struct {
	client *redis.Client
	stream string
}

func NewRedisTickStreamSink(client *redis.Client, stream string) *RedisTickStreamSink {
	if stream == "" {
		stream = "kalency:v1:stream:ticks"
	}
	return &RedisTickStreamSink{client: client, stream: stream}
}

func (s *RedisTickStreamSink) PublishTick(ctx context.Context, tick sim.Tick) error {
	ts := tick.TS
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	_, err := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: s.stream,
		Values: map[string]any{
			"symbol": tick.Symbol,
			"price":  strconv.FormatFloat(tick.Price, 'f', -1, 64),
			"delta":  strconv.FormatFloat(tick.Delta, 'f', -1, 64),
			"ts":     ts.Format(time.RFC3339Nano),
		},
	}).Result()
	return err
}
