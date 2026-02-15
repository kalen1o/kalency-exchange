package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/candle-aggregator/internal/candle"
)

type RedisCandleStore struct {
	client redis.UniversalClient
	prefix string
}

func NewRedisCandleStore(client redis.UniversalClient, prefix string) *RedisCandleStore {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "v1"
	}
	return &RedisCandleStore{client: client, prefix: prefix}
}

func (s *RedisCandleStore) UpsertCandle(ctx context.Context, point candle.Candle, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = candle.DefaultCandleTTL
	}

	key := s.candleKey(point)
	args := []any{
		point.Symbol,
		point.Timeframe,
		point.BucketStart.UTC().Format(time.RFC3339),
		strconv.FormatFloat(point.Open, 'f', -1, 64),
		strconv.FormatFloat(point.High, 'f', -1, 64),
		strconv.FormatFloat(point.Low, 'f', -1, 64),
		strconv.FormatFloat(point.Close, 'f', -1, 64),
		strconv.FormatFloat(point.Volume, 'f', -1, 64),
		strconv.FormatInt(int64(ttl.Seconds()), 10),
	}

	return upsertCandleScript.Run(ctx, s.client, []string{key}, args...).Err()
}

func (s *RedisCandleStore) candleKey(point candle.Candle) string {
	return fmt.Sprintf(
		"%s:candle:%s:%s:%s",
		s.prefix,
		point.Symbol,
		point.Timeframe,
		point.BucketStart.UTC().Format(time.RFC3339),
	)
}

var upsertCandleScript = redis.NewScript(`
local key = KEYS[1]

local symbol = ARGV[1]
local timeframe = ARGV[2]
local bucket_start = ARGV[3]
local open_in = ARGV[4]
local high_in = ARGV[5]
local low_in = ARGV[6]
local close_in = ARGV[7]
local volume_add = tonumber(ARGV[8])
local ttl_seconds = tonumber(ARGV[9])

local current = redis.call('HMGET', key, 'open', 'high', 'low', 'volume')
local open_cur = current[1]
local high_cur = current[2]
local low_cur = current[3]
local volume_cur = tonumber(current[4] or '0')

if (not open_cur) then
  open_cur = open_in
end

if (not high_cur) or (tonumber(high_in) > tonumber(high_cur)) then
  high_cur = high_in
end

if (not low_cur) or (tonumber(low_in) < tonumber(low_cur)) then
  low_cur = low_in
end

local volume_next = volume_cur + volume_add

redis.call('HSET', key,
  'symbol', symbol,
  'timeframe', timeframe,
  'bucket_start', bucket_start,
  'open', open_cur,
  'high', high_cur,
  'low', low_cur,
  'close', close_in,
  'volume', tostring(volume_next)
)
redis.call('EXPIRE', key, ttl_seconds)
return 1
`)
