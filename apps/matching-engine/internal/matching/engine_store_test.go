package matching

import (
	"context"
	"sort"
	"testing"
	"time"
)

type fakeOpenOrdersStore struct {
	data map[string][]Order
}

func newFakeOpenOrdersStore() *fakeOpenOrdersStore {
	return &fakeOpenOrdersStore{data: make(map[string][]Order)}
}

func (f *fakeOpenOrdersStore) SetUserOrders(_ context.Context, userID string, orders []Order) error {
	copied := make([]Order, len(orders))
	copy(copied, orders)
	f.data[userID] = copied
	return nil
}

func (f *fakeOpenOrdersStore) GetUserOrders(_ context.Context, userID string) ([]Order, bool, error) {
	orders, ok := f.data[userID]
	if !ok {
		return nil, false, nil
	}
	copied := make([]Order, len(orders))
	copy(copied, orders)
	return copied, true, nil
}

func TestEngineSyncsStoreForMakersAndTaker(t *testing.T) {
	store := newFakeOpenOrdersStore()
	engine := NewEngineWithStore(store)
	engine.FundWallet("seller1", "BTC", 5)
	engine.FundWallet("seller2", "BTC", 5)

	_, err := engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "s-1",
		UserID:        "seller1",
		Symbol:        "BTC-USD",
		Side:          SideSell,
		Type:          OrderTypeLimit,
		Price:         100,
		Qty:           5,
	})
	if err != nil {
		t.Fatalf("seed order 1 failed: %v", err)
	}

	_, err = engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "s-2",
		UserID:        "seller2",
		Symbol:        "BTC-USD",
		Side:          SideSell,
		Type:          OrderTypeLimit,
		Price:         100,
		Qty:           5,
	})
	if err != nil {
		t.Fatalf("seed order 2 failed: %v", err)
	}

	_, err = engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "b-1",
		UserID:        "buyer1",
		Symbol:        "BTC-USD",
		Side:          SideBuy,
		Type:          OrderTypeMarket,
		Qty:           7,
	})
	if err != nil {
		t.Fatalf("market order failed: %v", err)
	}

	seller1Orders, ok := store.data["seller1"]
	if !ok {
		t.Fatal("expected seller1 to be synced to store")
	}
	if len(seller1Orders) != 0 {
		t.Fatalf("expected seller1 to have 0 open orders in store, got %d", len(seller1Orders))
	}

	seller2Orders, ok := store.data["seller2"]
	if !ok {
		t.Fatal("expected seller2 to be synced to store")
	}
	if len(seller2Orders) != 1 {
		t.Fatalf("expected seller2 to have 1 open order in store, got %d", len(seller2Orders))
	}
	if seller2Orders[0].RemainingQty != 3 {
		t.Fatalf("expected seller2 remaining qty 3 in store, got %d", seller2Orders[0].RemainingQty)
	}

	buyerOrders, ok := store.data["buyer1"]
	if !ok {
		t.Fatal("expected buyer1 to be synced to store")
	}
	if len(buyerOrders) != 0 {
		t.Fatalf("expected buyer1 to have 0 open orders in store, got %d", len(buyerOrders))
	}
}

func TestOpenOrdersReadsFromStoreWhenAvailable(t *testing.T) {
	store := newFakeOpenOrdersStore()
	store.data["u1"] = []Order{{
		OrderID:      "ord-redis-1",
		UserID:       "u1",
		Symbol:       "BTC-USD",
		Side:         SideBuy,
		Type:         OrderTypeLimit,
		Price:        100,
		Qty:          10,
		RemainingQty: 10,
		CreatedAt:    time.Unix(1, 0).UTC(),
	}}
	engine := NewEngineWithStore(store)

	got := engine.OpenOrders("u1")
	if len(got) != 1 {
		t.Fatalf("expected 1 order from store, got %d", len(got))
	}
	if got[0].OrderID != "ord-redis-1" {
		t.Fatalf("expected order id ord-redis-1, got %s", got[0].OrderID)
	}
}

func TestOpenOrdersSortsChronologicallyFromStore(t *testing.T) {
	store := newFakeOpenOrdersStore()
	store.data["u1"] = []Order{
		{OrderID: "ord-2", UserID: "u1", CreatedAt: time.Unix(2, 0).UTC()},
		{OrderID: "ord-1", UserID: "u1", CreatedAt: time.Unix(1, 0).UTC()},
	}
	engine := NewEngineWithStore(store)

	got := engine.OpenOrders("u1")
	if len(got) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(got))
	}
	if !sort.SliceIsSorted(got, func(i, j int) bool { return got[i].CreatedAt.Before(got[j].CreatedAt) }) {
		t.Fatal("expected open orders to be returned in chronological order")
	}
}
