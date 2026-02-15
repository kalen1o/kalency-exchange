package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisTickStreamSourceReadParsesTicks(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	stream := "kalency:v1:stream:ticks"
	msg1ID, err := client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: stream,
		Values: map[string]any{
			"symbol": "BTC-USD",
			"price":  "101.25",
			"ts":     "2026-02-14T12:34:56Z",
		},
	}).Result()
	if err != nil {
		t.Fatalf("xadd msg1 failed: %v", err)
	}

	msg2ID, err := client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: stream,
		Values: map[string]any{
			"symbol": "BTC-USD",
			"price":  "102.50",
			"volume": "3.5",
			"ts":     "2026-02-14T12:34:57Z",
		},
	}).Result()
	if err != nil {
		t.Fatalf("xadd msg2 failed: %v", err)
	}

	source := NewRedisTickStreamSource(client, stream)
	ticks, lastID, err := source.Read(context.Background(), "0-0", 10, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if len(ticks) != 2 {
		t.Fatalf("expected 2 ticks, got %d", len(ticks))
	}
	if ticks[0].Volume != 1 {
		t.Fatalf("expected default volume 1, got %f", ticks[0].Volume)
	}
	if ticks[1].Volume != 3.5 {
		t.Fatalf("expected volume 3.5, got %f", ticks[1].Volume)
	}
	if ticks[0].Price != 101.25 || ticks[1].Price != 102.5 {
		t.Fatalf("unexpected prices: %f %f", ticks[0].Price, ticks[1].Price)
	}
	if lastID != msg2ID {
		t.Fatalf("expected last id %s, got %s (first id was %s)", msg2ID, lastID, msg1ID)
	}
}
