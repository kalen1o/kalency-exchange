package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/ledger-writer/internal/ledger"
)

type RedisExecutionStreamSource struct {
	client redis.UniversalClient
	stream string
}

func NewRedisExecutionStreamSource(client redis.UniversalClient, stream string) *RedisExecutionStreamSource {
	stream = strings.TrimSpace(stream)
	if stream == "" {
		stream = "kalency:v1:stream:executions"
	}
	return &RedisExecutionStreamSource{client: client, stream: stream}
}

func (s *RedisExecutionStreamSource) Read(ctx context.Context, lastID string, count int, block time.Duration) ([]ledger.ExecutionEvent, string, error) {
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
		return []ledger.ExecutionEvent{}, lastID, nil
	}
	if err != nil {
		return nil, lastID, err
	}

	result := make([]ledger.ExecutionEvent, 0)
	nextID := lastID
	for _, stream := range streamData {
		for _, message := range stream.Messages {
			nextID = message.ID
			event, decodeErr := decodeExecution(message.Values)
			if decodeErr != nil {
				continue
			}
			result = append(result, event)
		}
	}

	return result, nextID, nil
}

func decodeExecution(values map[string]any) (ledger.ExecutionEvent, error) {
	tradeID := strings.TrimSpace(fmt.Sprint(values["trade_id"]))
	symbol := strings.TrimSpace(fmt.Sprint(values["symbol"]))
	if tradeID == "" || symbol == "" {
		return ledger.ExecutionEvent{}, fmt.Errorf("missing identifiers")
	}

	price, err := parseFloat(values["price"])
	if err != nil || price <= 0 {
		return ledger.ExecutionEvent{}, fmt.Errorf("invalid price")
	}
	qty, err := parseFloat(values["qty"])
	if err != nil || qty <= 0 {
		return ledger.ExecutionEvent{}, fmt.Errorf("invalid qty")
	}

	makerUserID := strings.TrimSpace(fmt.Sprint(values["maker_user_id"]))
	takerUserID := strings.TrimSpace(fmt.Sprint(values["taker_user_id"]))

	ts := time.Now().UTC()
	if raw, ok := values["ts"]; ok {
		if parsed, parseErr := time.Parse(time.RFC3339Nano, fmt.Sprint(raw)); parseErr == nil {
			ts = parsed.UTC()
		}
	}

	return ledger.ExecutionEvent{
		TradeID:    tradeID,
		Symbol:     symbol,
		BuyUserID:  takerUserID,
		SellUserID: makerUserID,
		Price:      price,
		Qty:        qty,
		ExecutedAt: ts,
	}, nil
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
