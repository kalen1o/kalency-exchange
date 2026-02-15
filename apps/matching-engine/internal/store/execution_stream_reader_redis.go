package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"kalency/apps/matching-engine/internal/matching"
)

type RedisExecutionStreamReader struct {
	client redis.UniversalClient
	stream string
}

func NewRedisExecutionStreamReader(client redis.UniversalClient, stream string) *RedisExecutionStreamReader {
	if stream == "" {
		stream = "kalency:v1:stream:executions"
	}
	return &RedisExecutionStreamReader{client: client, stream: stream}
}

func (r *RedisExecutionStreamReader) ListExecutions(symbol string, limit int) ([]matching.Execution, error) {
	if limit <= 0 {
		limit = 100
	}

	fetchCount := int64(limit * 20)
	if fetchCount < 100 {
		fetchCount = 100
	}

	entries, err := r.client.XRevRangeN(context.Background(), r.stream, "+", "-", fetchCount).Result()
	if err == redis.Nil {
		return []matching.Execution{}, nil
	}
	if err != nil {
		return nil, err
	}

	filtered := make([]matching.Execution, 0, limit)
	for _, entry := range entries {
		execution, err := decodeExecution(entry.Values)
		if err != nil {
			continue
		}
		if execution.Symbol != symbol {
			continue
		}
		filtered = append(filtered, execution)
		if len(filtered) == limit {
			break
		}
	}

	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}
	return filtered, nil
}

func decodeExecution(values map[string]any) (matching.Execution, error) {
	price, err := parseInt64(values["price"])
	if err != nil {
		return matching.Execution{}, err
	}
	qty, err := parseInt64(values["qty"])
	if err != nil {
		return matching.Execution{}, err
	}
	tsValue := fmt.Sprint(values["ts"])
	ts, err := time.Parse(time.RFC3339Nano, tsValue)
	if err != nil {
		ts = time.Time{}
	}

	return matching.Execution{
		TradeID:      fmt.Sprint(values["trade_id"]),
		Symbol:       fmt.Sprint(values["symbol"]),
		Price:        price,
		Qty:          qty,
		MakerOrderID: fmt.Sprint(values["maker_order_id"]),
		MakerUserID:  fmt.Sprint(values["maker_user_id"]),
		TakerOrderID: fmt.Sprint(values["taker_order_id"]),
		TakerUserID:  fmt.Sprint(values["taker_user_id"]),
		TS:           ts,
	}, nil
}

func parseInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return strconv.ParseInt(fmt.Sprint(v), 10, 64)
	}
}
