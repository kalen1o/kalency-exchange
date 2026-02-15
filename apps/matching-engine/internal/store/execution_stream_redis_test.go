package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"kalency/apps/matching-engine/internal/matching"
)

func TestRedisExecutionStreamSinkPublishExecution(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	stream := "kalency:v1:stream:executions"
	sink := NewRedisExecutionStreamSink(client, stream)

	exec := matching.Execution{
		TradeID:      "trd-1",
		Symbol:       "BTC-USD",
		Price:        101,
		Qty:          3,
		MakerOrderID: "ord-1",
		MakerUserID:  "seller1",
		TakerOrderID: "ord-2",
		TakerUserID:  "buyer1",
		TS:           time.Unix(10, 0).UTC(),
	}

	if err := sink.PublishExecution(context.Background(), exec); err != nil {
		t.Fatalf("publish execution failed: %v", err)
	}

	messages, err := client.XRange(context.Background(), stream, "-", "+").Result()
	if err != nil {
		t.Fatalf("xrange failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 stream message, got %d", len(messages))
	}

	values := messages[0].Values
	if got := fmt.Sprint(values["trade_id"]); got != "trd-1" {
		t.Fatalf("expected trade_id trd-1, got %s", got)
	}
	if got := fmt.Sprint(values["symbol"]); got != "BTC-USD" {
		t.Fatalf("expected symbol BTC-USD, got %s", got)
	}
	if got := fmt.Sprint(values["qty"]); got != "3" {
		t.Fatalf("expected qty 3, got %s", got)
	}
}
