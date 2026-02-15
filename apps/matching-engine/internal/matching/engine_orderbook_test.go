package matching

import "testing"

func TestOrderBookSnapshotAggregatesAndLimitsDepth(t *testing.T) {
	engine := NewEngine()
	engine.FundWallet("seller1", "BTC", 10)
	engine.FundWallet("seller2", "BTC", 10)

	_, err := engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "s-1",
		UserID:        "seller1",
		Symbol:        "BTC-USD",
		Side:          SideSell,
		Type:          OrderTypeLimit,
		Price:         110,
		Qty:           2,
	})
	if err != nil {
		t.Fatalf("place sell order s-1 failed: %v", err)
	}

	_, err = engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "s-2",
		UserID:        "seller2",
		Symbol:        "BTC-USD",
		Side:          SideSell,
		Type:          OrderTypeLimit,
		Price:         110,
		Qty:           3,
	})
	if err != nil {
		t.Fatalf("place sell order s-2 failed: %v", err)
	}

	_, err = engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "s-3",
		UserID:        "seller1",
		Symbol:        "BTC-USD",
		Side:          SideSell,
		Type:          OrderTypeLimit,
		Price:         111,
		Qty:           4,
	})
	if err != nil {
		t.Fatalf("place sell order s-3 failed: %v", err)
	}

	_, err = engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "b-1",
		UserID:        "buyer1",
		Symbol:        "BTC-USD",
		Side:          SideBuy,
		Type:          OrderTypeLimit,
		Price:         100,
		Qty:           1,
	})
	if err != nil {
		t.Fatalf("place buy order b-1 failed: %v", err)
	}

	_, err = engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "b-2",
		UserID:        "buyer2",
		Symbol:        "BTC-USD",
		Side:          SideBuy,
		Type:          OrderTypeLimit,
		Price:         100,
		Qty:           2,
	})
	if err != nil {
		t.Fatalf("place buy order b-2 failed: %v", err)
	}

	_, err = engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "b-3",
		UserID:        "buyer3",
		Symbol:        "BTC-USD",
		Side:          SideBuy,
		Type:          OrderTypeLimit,
		Price:         99,
		Qty:           7,
	})
	if err != nil {
		t.Fatalf("place buy order b-3 failed: %v", err)
	}

	snapshot := engine.OrderBookSnapshot("BTC-USD", 1)

	if len(snapshot.Bids) != 1 {
		t.Fatalf("expected 1 bid level, got %d", len(snapshot.Bids))
	}
	if len(snapshot.Asks) != 1 {
		t.Fatalf("expected 1 ask level, got %d", len(snapshot.Asks))
	}

	if snapshot.Bids[0].Price != 100 {
		t.Fatalf("expected top bid 100, got %d", snapshot.Bids[0].Price)
	}
	if snapshot.Bids[0].Qty != 3 {
		t.Fatalf("expected top bid qty 3, got %d", snapshot.Bids[0].Qty)
	}
	if snapshot.Bids[0].Orders != 2 {
		t.Fatalf("expected top bid order count 2, got %d", snapshot.Bids[0].Orders)
	}

	if snapshot.Asks[0].Price != 110 {
		t.Fatalf("expected top ask 110, got %d", snapshot.Asks[0].Price)
	}
	if snapshot.Asks[0].Qty != 5 {
		t.Fatalf("expected top ask qty 5, got %d", snapshot.Asks[0].Qty)
	}
	if snapshot.Asks[0].Orders != 2 {
		t.Fatalf("expected top ask order count 2, got %d", snapshot.Asks[0].Orders)
	}
}
