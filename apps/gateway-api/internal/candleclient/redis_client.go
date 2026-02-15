package candleclient

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/gateway-api/internal/contracts"
)

type RedisClient struct {
	client redis.UniversalClient
	prefix string
}

func NewRedisClient(client redis.UniversalClient, prefix string) *RedisClient {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "v1"
	}
	return &RedisClient{client: client, prefix: prefix}
}

func (r *RedisClient) ListCandles(symbol, timeframe string, from, to time.Time) ([]contracts.Candle, error) {
	symbol = strings.TrimSpace(symbol)
	timeframe = strings.TrimSpace(timeframe)
	if symbol == "" || timeframe == "" {
		return []contracts.Candle{}, nil
	}

	pattern := fmt.Sprintf("%s:candle:%s:%s:*", r.prefix, symbol, timeframe)
	ctx := context.Background()

	keys, err := scanAllKeys(ctx, r.client, pattern)
	if err != nil {
		return nil, err
	}

	candles := make([]contracts.Candle, 0, len(keys))
	for _, key := range keys {
		values, err := r.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		candle, ok := decodeCandle(values)
		if !ok {
			continue
		}
		if !from.IsZero() && candle.BucketStart.Before(from) {
			continue
		}
		if !to.IsZero() && candle.BucketStart.After(to) {
			continue
		}
		candles = append(candles, candle)
	}

	sort.Slice(candles, func(i, j int) bool {
		return candles[i].BucketStart.Before(candles[j].BucketStart)
	})
	return candles, nil
}

func scanAllKeys(ctx context.Context, client redis.UniversalClient, pattern string) ([]string, error) {
	cursor := uint64(0)
	all := make([]string, 0)
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, pattern, 1000).Result()
		if err != nil {
			return nil, err
		}
		all = append(all, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return all, nil
}

func decodeCandle(values map[string]string) (contracts.Candle, bool) {
	symbol := strings.TrimSpace(values["symbol"])
	timeframe := strings.TrimSpace(values["timeframe"])
	bucketRaw := strings.TrimSpace(values["bucket_start"])
	if symbol == "" || timeframe == "" || bucketRaw == "" {
		return contracts.Candle{}, false
	}

	bucketStart, err := time.Parse(time.RFC3339, bucketRaw)
	if err != nil {
		bucketStart, err = time.Parse(time.RFC3339Nano, bucketRaw)
		if err != nil {
			return contracts.Candle{}, false
		}
	}

	open, ok := parseFloat(values["open"])
	if !ok {
		return contracts.Candle{}, false
	}
	high, ok := parseFloat(values["high"])
	if !ok {
		return contracts.Candle{}, false
	}
	low, ok := parseFloat(values["low"])
	if !ok {
		return contracts.Candle{}, false
	}
	closeValue, ok := parseFloat(values["close"])
	if !ok {
		return contracts.Candle{}, false
	}
	volume, ok := parseFloat(values["volume"])
	if !ok {
		return contracts.Candle{}, false
	}

	return contracts.Candle{
		Symbol:      symbol,
		Timeframe:   timeframe,
		BucketStart: bucketStart.UTC(),
		Open:        open,
		High:        high,
		Low:         low,
		Close:       closeValue,
		Volume:      volume,
	}, true
}

func parseFloat(raw string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, false
	}
	return value, true
}
