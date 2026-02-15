package candle

import (
	"context"
	"testing"
	"time"
)

type recordedUpsert struct {
	Candle Candle
	TTL    time.Duration
}

type recordingStore struct {
	calls []recordedUpsert
}

func (r *recordingStore) UpsertCandle(_ context.Context, candle Candle, ttl time.Duration) error {
	r.calls = append(r.calls, recordedUpsert{Candle: candle, TTL: ttl})
	return nil
}

func TestServiceProcessTickWritesAllTimeframes(t *testing.T) {
	store := &recordingStore{}
	svc := NewService(store, Config{})

	tickTime := time.Date(2026, 2, 14, 12, 34, 56, 789000000, time.UTC)
	tick := Tick{Symbol: "BTC-USD", Price: 101.25, Volume: 2.5, TS: tickTime}
	if err := svc.ProcessTick(context.Background(), tick); err != nil {
		t.Fatalf("process tick failed: %v", err)
	}

	if len(store.calls) != 5 {
		t.Fatalf("expected 5 upserts, got %d", len(store.calls))
	}

	expectedBuckets := map[string]time.Time{
		"1s": time.Date(2026, 2, 14, 12, 34, 56, 0, time.UTC),
		"5s": time.Date(2026, 2, 14, 12, 34, 55, 0, time.UTC),
		"1m": time.Date(2026, 2, 14, 12, 34, 0, 0, time.UTC),
		"5m": time.Date(2026, 2, 14, 12, 30, 0, 0, time.UTC),
		"1h": time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC),
	}

	for _, call := range store.calls {
		expectedBucket, ok := expectedBuckets[call.Candle.Timeframe]
		if !ok {
			t.Fatalf("unexpected timeframe %s", call.Candle.Timeframe)
		}
		if !call.Candle.BucketStart.Equal(expectedBucket) {
			t.Fatalf("expected bucket %s for %s, got %s", expectedBucket, call.Candle.Timeframe, call.Candle.BucketStart)
		}
		if call.Candle.Symbol != "BTC-USD" {
			t.Fatalf("expected symbol BTC-USD, got %s", call.Candle.Symbol)
		}
		if call.Candle.Open != 101.25 || call.Candle.High != 101.25 || call.Candle.Low != 101.25 || call.Candle.Close != 101.25 {
			t.Fatalf("expected OHLC all 101.25, got o=%f h=%f l=%f c=%f", call.Candle.Open, call.Candle.High, call.Candle.Low, call.Candle.Close)
		}
		if call.Candle.Volume != 2.5 {
			t.Fatalf("expected volume 2.5, got %f", call.Candle.Volume)
		}
		if call.TTL != 30*24*time.Hour {
			t.Fatalf("expected 30d ttl, got %s", call.TTL)
		}
	}
}
