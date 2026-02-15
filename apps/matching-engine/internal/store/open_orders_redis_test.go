package store

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"kalency/apps/matching-engine/internal/matching"
)

func TestRedisOpenOrdersStoreSetAndGet(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	store := NewRedisOpenOrdersStore(client, "kalency:v1")
	ctx := context.Background()

	expected := []matching.Order{{
		OrderID:      "ord-1",
		UserID:       "u1",
		Symbol:       "BTC-USD",
		Side:         matching.SideBuy,
		Type:         matching.OrderTypeLimit,
		Price:        100,
		Qty:          10,
		RemainingQty: 10,
		CreatedAt:    time.Unix(1, 0).UTC(),
	}}

	if err := store.SetUserOrders(ctx, "u1", expected); err != nil {
		t.Fatalf("set user orders failed: %v", err)
	}

	got, found, err := store.GetUserOrders(ctx, "u1")
	if err != nil {
		t.Fatalf("get user orders failed: %v", err)
	}
	if !found {
		t.Fatal("expected user orders to be found")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 order, got %d", len(got))
	}
	if got[0].OrderID != expected[0].OrderID {
		t.Fatalf("expected order id %s, got %s", expected[0].OrderID, got[0].OrderID)
	}
}

func TestRedisOpenOrdersStoreGetMissingReturnsNotFound(t *testing.T) {
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mini.Close()

	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	store := NewRedisOpenOrdersStore(client, "kalency:v1")
	ctx := context.Background()

	got, found, err := store.GetUserOrders(ctx, "does-not-exist")
	if err != nil {
		t.Fatalf("get user orders failed: %v", err)
	}
	if found {
		t.Fatal("expected missing user to return found=false")
	}
	if len(got) != 0 {
		t.Fatalf("expected no orders, got %d", len(got))
	}
}
