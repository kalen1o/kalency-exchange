package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"kalency/apps/matching-engine/internal/matching"
)

func TestRedisExecutionStreamReaderListExecutionsBySymbol(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	stream := "kalency:v1:stream:executions"
	sink := NewRedisExecutionStreamSink(client, stream)
	reader := NewRedisExecutionStreamReader(client, stream)

	if err := sink.PublishExecution(context.Background(), matching.Execution{
		TradeID: "trd-btc-1", Symbol: "BTC-USD", Price: 100, Qty: 1,
		MakerOrderID: "m1", MakerUserID: "seller1", TakerOrderID: "t1", TakerUserID: "buyer1", TS: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("publish BTC execution failed: %v", err)
	}
	if err := sink.PublishExecution(context.Background(), matching.Execution{
		TradeID: "trd-eth-1", Symbol: "ETH-USD", Price: 200, Qty: 2,
		MakerOrderID: "m2", MakerUserID: "seller2", TakerOrderID: "t2", TakerUserID: "buyer2", TS: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("publish ETH execution failed: %v", err)
	}

	trades, err := reader.ListExecutions("BTC-USD", 10)
	if err != nil {
		t.Fatalf("list executions failed: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 BTC trade, got %d", len(trades))
	}
	if trades[0].TradeID != "trd-btc-1" {
		t.Fatalf("expected trade id trd-btc-1, got %s", trades[0].TradeID)
	}
}
