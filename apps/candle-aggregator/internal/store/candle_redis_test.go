package store

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"kalency/apps/candle-aggregator/internal/candle"
)

func TestRedisCandleStoreUpsertCandleAggregatesOHLCV(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	store := NewRedisCandleStore(client, "v1")
	bucket := time.Date(2026, 2, 14, 12, 34, 0, 0, time.UTC)

	first := candle.Candle{Symbol: "BTC-USD", Timeframe: "1m", BucketStart: bucket, Open: 100, High: 100, Low: 100, Close: 100, Volume: 2}
	second := candle.Candle{Symbol: "BTC-USD", Timeframe: "1m", BucketStart: bucket, Open: 110, High: 110, Low: 110, Close: 110, Volume: 3}

	if err := store.UpsertCandle(context.Background(), first, 30*24*time.Hour); err != nil {
		t.Fatalf("first upsert failed: %v", err)
	}
	if err := store.UpsertCandle(context.Background(), second, 30*24*time.Hour); err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}

	key := "v1:candle:BTC-USD:1m:2026-02-14T12:34:00Z"
	values, err := client.HGetAll(context.Background(), key).Result()
	if err != nil {
		t.Fatalf("hgetall failed: %v", err)
	}
	if len(values) == 0 {
		t.Fatalf("expected candle hash at key %s", key)
	}

	assertFloatField(t, values, "open", 100)
	assertFloatField(t, values, "high", 110)
	assertFloatField(t, values, "low", 100)
	assertFloatField(t, values, "close", 110)
	assertFloatField(t, values, "volume", 5)

	ttl := mini.TTL(key)
	if ttl <= 0 {
		t.Fatalf("expected ttl > 0, got %s", ttl)
	}
}

func assertFloatField(t *testing.T, values map[string]string, field string, expected float64) {
	t.Helper()
	raw, ok := values[field]
	if !ok {
		t.Fatalf("missing field %s", field)
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		t.Fatalf("parse field %s: %v", field, err)
	}
	if parsed != expected {
		t.Fatalf("expected %s=%f, got %f", field, expected, parsed)
	}
}
