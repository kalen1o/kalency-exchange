package candle

import (
	"context"
	"errors"
	"strings"
	"time"
)

const DefaultCandleTTL = 30 * 24 * time.Hour

type CandleStore interface {
	UpsertCandle(ctx context.Context, candle Candle, ttl time.Duration) error
}

type Config struct {
	Timeframes []string
	TTL        time.Duration
}

type timeframeDef struct {
	Name     string
	Duration time.Duration
}

var supportedTimeframes = map[string]time.Duration{
	"1s": time.Second,
	"5s": 5 * time.Second,
	"1m": time.Minute,
	"5m": 5 * time.Minute,
	"1h": time.Hour,
}

var defaultTimeframeOrder = []string{"1s", "5s", "1m", "5m", "1h"}

type Service struct {
	store      CandleStore
	timeframes []timeframeDef
	ttl        time.Duration
}

func NewService(store CandleStore, cfg Config) *Service {
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = DefaultCandleTTL
	}

	timeframes := parseTimeframes(cfg.Timeframes)
	if len(timeframes) == 0 {
		timeframes = parseTimeframes(defaultTimeframeOrder)
	}

	return &Service{store: store, timeframes: timeframes, ttl: ttl}
}

func (s *Service) ProcessTick(ctx context.Context, tick Tick) error {
	if s.store == nil {
		return errors.New("candle store is required")
	}

	tick.Symbol = strings.TrimSpace(tick.Symbol)
	if tick.Symbol == "" {
		return errors.New("tick symbol is required")
	}
	if tick.Price <= 0 {
		return errors.New("tick price must be positive")
	}
	if tick.Volume <= 0 {
		tick.Volume = 1
	}
	if tick.TS.IsZero() {
		tick.TS = time.Now().UTC()
	}

	for _, timeframe := range s.timeframes {
		bucketStart := bucketStart(tick.TS, timeframe.Duration)
		point := Candle{
			Symbol:      tick.Symbol,
			Timeframe:   timeframe.Name,
			BucketStart: bucketStart,
			Open:        tick.Price,
			High:        tick.Price,
			Low:         tick.Price,
			Close:       tick.Price,
			Volume:      tick.Volume,
		}
		if err := s.store.UpsertCandle(ctx, point, s.ttl); err != nil {
			return err
		}
	}
	return nil
}

func parseTimeframes(raw []string) []timeframeDef {
	result := make([]timeframeDef, 0, len(raw))
	seen := map[string]struct{}{}
	for _, value := range raw {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		duration, ok := supportedTimeframes[normalized]
		if !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, timeframeDef{Name: normalized, Duration: duration})
	}
	return result
}

func bucketStart(ts time.Time, duration time.Duration) time.Time {
	ts = ts.UTC()
	unix := ts.Unix()
	bucket := unix - (unix % int64(duration.Seconds()))
	return time.Unix(bucket, 0).UTC()
}
