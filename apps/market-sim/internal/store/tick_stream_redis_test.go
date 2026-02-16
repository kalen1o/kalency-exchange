package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"kalency/apps/market-sim/internal/sim"
)

func TestRedisTickStreamSinkPublishTick(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	stream := "kalency:v1:stream:ticks"
	sink := NewRedisTickStreamSink(client, stream)

	tick := sim.Tick{
		Symbol: "BTC-USD",
		Price:  101.25,
		Delta:  0.0125,
		Volume: 1.75,
		TS:     time.Unix(10, 0).UTC(),
	}

	if err := sink.PublishTick(context.Background(), tick); err != nil {
		t.Fatalf("publish tick failed: %v", err)
	}

	messages, err := client.XRange(context.Background(), stream, "-", "+").Result()
	if err != nil {
		t.Fatalf("xrange failed: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 stream message, got %d", len(messages))
	}

	values := messages[0].Values
	if got := fmt.Sprint(values["symbol"]); got != "BTC-USD" {
		t.Fatalf("expected symbol BTC-USD, got %s", got)
	}
	if got := fmt.Sprint(values["price"]); got != "101.25" {
		t.Fatalf("expected price 101.25, got %s", got)
	}
	if got := fmt.Sprint(values["volume"]); got != "1.75" {
		t.Fatalf("expected volume 1.75, got %s", got)
	}
}
