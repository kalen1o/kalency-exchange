package candleclient

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisClientListCandlesFiltersAndSorts(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	seed := func(key string, values map[string]string) {
		if err := client.HSet(context.Background(), key, values).Err(); err != nil {
			t.Fatalf("seed hset %s failed: %v", key, err)
		}
	}

	seed("v1:candle:BTC-USD:1m:2026-02-15T00:00:00Z", map[string]string{
		"symbol": "BTC-USD", "timeframe": "1m", "bucket_start": "2026-02-15T00:00:00Z",
		"open": "100", "high": "101", "low": "99", "close": "100.5", "volume": "10",
	})
	seed("v1:candle:BTC-USD:1m:2026-02-15T00:01:00Z", map[string]string{
		"symbol": "BTC-USD", "timeframe": "1m", "bucket_start": "2026-02-15T00:01:00Z",
		"open": "100.5", "high": "102", "low": "100", "close": "101", "volume": "12",
	})

	rc := NewRedisClient(client, "v1")
	from := time.Date(2026, 2, 15, 0, 0, 30, 0, time.UTC)
	to := time.Date(2026, 2, 15, 0, 2, 0, 0, time.UTC)

	candles, err := rc.ListCandles("BTC-USD", "1m", from, to)
	if err != nil {
		t.Fatalf("list candles failed: %v", err)
	}
	if len(candles) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(candles))
	}
	if !candles[0].BucketStart.Equal(time.Date(2026, 2, 15, 0, 1, 0, 0, time.UTC)) {
		t.Fatalf("unexpected bucket start: %s", candles[0].BucketStart)
	}
}
