package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"kalency/apps/matching-engine/internal/matching"
)

type RedisOpenOrdersStore struct {
	client redis.UniversalClient
	prefix string
}

func NewRedisOpenOrdersStore(client redis.UniversalClient, prefix string) *RedisOpenOrdersStore {
	if prefix == "" {
		prefix = "kalency:v1"
	}
	return &RedisOpenOrdersStore{client: client, prefix: prefix}
}

func (s *RedisOpenOrdersStore) SetUserOrders(ctx context.Context, userID string, orders []matching.Order) error {
	if len(orders) == 0 {
		return s.client.Del(ctx, s.key(userID)).Err()
	}

	payload, err := json.Marshal(orders)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, s.key(userID), payload, 0).Err()
}

func (s *RedisOpenOrdersStore) GetUserOrders(ctx context.Context, userID string) ([]matching.Order, bool, error) {
	payload, err := s.client.Get(ctx, s.key(userID)).Bytes()
	if err == redis.Nil {
		return []matching.Order{}, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var orders []matching.Order
	if err := json.Unmarshal(payload, &orders); err != nil {
		return nil, false, err
	}
	return orders, true, nil
}

func (s *RedisOpenOrdersStore) key(userID string) string {
	return fmt.Sprintf("%s:orders:open:%s", s.prefix, userID)
}
