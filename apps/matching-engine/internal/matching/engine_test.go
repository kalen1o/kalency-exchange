package matching

import "testing"

func TestPlaceLimitOrderStoresOpenOrder(t *testing.T) {
	engine := NewEngine()

	ack, err := engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "c-1",
		UserID:        "u1",
		Symbol:        "BTC-USD",
		Side:          SideBuy,
		Type:          OrderTypeLimit,
		Price:         100,
		Qty:           10,
	})
	if err != nil {
		t.Fatalf("place order returned error: %v", err)
	}

	if ack.Status != OrderStatusAccepted {
		t.Fatalf("expected %s status, got %s", OrderStatusAccepted, ack.Status)
	}
	if ack.RemainingQty != 10 {
		t.Fatalf("expected remaining qty 10, got %d", ack.RemainingQty)
	}

	open := engine.OpenOrders("u1")
	if len(open) != 1 {
		t.Fatalf("expected 1 open order, got %d", len(open))
	}
	if open[0].OrderID != ack.OrderID {
		t.Fatalf("expected open order id %s, got %s", ack.OrderID, open[0].OrderID)
	}
}

func TestMarketOrderMatchesPriceTimePriority(t *testing.T) {
	engine := NewEngine()
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

	ack, err := engine.PlaceOrder(PlaceOrderRequest{
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

	if ack.Status != OrderStatusFilled {
		t.Fatalf("expected %s, got %s", OrderStatusFilled, ack.Status)
	}
	if ack.FilledQty != 7 {
		t.Fatalf("expected filled qty 7, got %d", ack.FilledQty)
	}
	if ack.AvgPrice != 100 {
		t.Fatalf("expected avg price 100, got %d", ack.AvgPrice)
	}

	if got := len(engine.OpenOrders("seller1")); got != 0 {
		t.Fatalf("expected seller1 to have 0 open orders, got %d", got)
	}

	seller2Open := engine.OpenOrders("seller2")
	if len(seller2Open) != 1 {
		t.Fatalf("expected seller2 to have 1 open order, got %d", len(seller2Open))
	}
	if seller2Open[0].RemainingQty != 3 {
		t.Fatalf("expected seller2 remaining qty 3, got %d", seller2Open[0].RemainingQty)
	}

	execs := engine.Executions("BTC-USD")
	if len(execs) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(execs))
	}
	if execs[0].MakerUserID != "seller1" || execs[0].Qty != 5 {
		t.Fatalf("expected first match seller1 qty=5, got maker=%s qty=%d", execs[0].MakerUserID, execs[0].Qty)
	}
	if execs[1].MakerUserID != "seller2" || execs[1].Qty != 2 {
		t.Fatalf("expected second match seller2 qty=2, got maker=%s qty=%d", execs[1].MakerUserID, execs[1].Qty)
	}
}
