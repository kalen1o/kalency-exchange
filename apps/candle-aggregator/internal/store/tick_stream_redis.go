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

type RedisTickStreamSource struct {
	client redis.UniversalClient
	stream string
}

func NewRedisTickStreamSource(client redis.UniversalClient, stream string) *RedisTickStreamSource {
	stream = strings.TrimSpace(stream)
	if stream == "" {
		stream = "kalency:v1:stream:ticks"
	}
	return &RedisTickStreamSource{client: client, stream: stream}
}

func (s *RedisTickStreamSource) Read(ctx context.Context, lastID string, count int, block time.Duration) ([]candle.Tick, string, error) {
	if strings.TrimSpace(lastID) == "" {
		lastID = "$"
	}
	if count <= 0 {
		count = 100
	}
	if block < 0 {
		block = 0
	}

	streamData, err := s.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{s.stream, lastID},
		Count:   int64(count),
		Block:   block,
	}).Result()
	if err == redis.Nil {
		return []candle.Tick{}, lastID, nil
	}
	if err != nil {
		return nil, lastID, err
	}

	result := make([]candle.Tick, 0)
	nextID := lastID
	for _, stream := range streamData {
		for _, message := range stream.Messages {
			nextID = message.ID
			tick, decodeErr := decodeTick(message.Values)
			if decodeErr != nil {
				continue
			}
			result = append(result, tick)
		}
	}

	return result, nextID, nil
}

func decodeTick(values map[string]any) (candle.Tick, error) {
	symbol := strings.TrimSpace(fmt.Sprint(values["symbol"]))
	if symbol == "" {
		return candle.Tick{}, fmt.Errorf("missing symbol")
	}

	price, err := parseFloat(values["price"])
	if err != nil || price <= 0 {
		return candle.Tick{}, fmt.Errorf("invalid price")
	}

	volume := 1.0
	if raw, ok := values["volume"]; ok {
		if parsed, parseErr := parseFloat(raw); parseErr == nil && parsed > 0 {
			volume = parsed
		}
	}
	if raw, ok := values["qty"]; ok {
		if parsed, parseErr := parseFloat(raw); parseErr == nil && parsed > 0 {
			volume = parsed
		}
	}

	ts := time.Now().UTC()
	if raw, ok := values["ts"]; ok {
		if parsed, parseErr := time.Parse(time.RFC3339Nano, fmt.Sprint(raw)); parseErr == nil {
			ts = parsed.UTC()
		}
	}

	return candle.Tick{Symbol: symbol, Price: price, Volume: volume, TS: ts}, nil
}

func parseFloat(value any) (float64, error) {
	switch typed := value.(type) {
	case float64:
		return typed, nil
	case float32:
		return float64(typed), nil
	case int:
		return float64(typed), nil
	case int64:
		return float64(typed), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(typed), 64)
	default:
		return strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(value)), 64)
	}
}
