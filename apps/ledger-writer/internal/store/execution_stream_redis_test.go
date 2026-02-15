package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisExecutionStreamSourceReadParsesEvents(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	stream := "kalency:v1:stream:executions"
	_, err = client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: stream,
		Values: map[string]any{
			"trade_id":      "trd-1",
			"symbol":        "BTC-USD",
			"price":         "101.25",
			"qty":           "2.5",
			"maker_user_id": "seller1",
			"taker_user_id": "buyer1",
			"ts":            "2026-02-15T00:00:00Z",
		},
	}).Result()
	if err != nil {
		t.Fatalf("xadd failed: %v", err)
	}

	source := NewRedisExecutionStreamSource(client, stream)
	events, _, err := source.Read(context.Background(), "0-0", 10, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].TradeID != "trd-1" {
		t.Fatalf("expected trade id trd-1, got %s", events[0].TradeID)
	}
	if events[0].BuyUserID != "buyer1" || events[0].SellUserID != "seller1" {
		t.Fatalf("unexpected user mapping buy=%s sell=%s", events[0].BuyUserID, events[0].SellUserID)
	}
}
