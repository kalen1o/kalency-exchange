package matching

import "testing"

func TestCancelOrderReleasesReservedQuoteBalance(t *testing.T) {
	engine := NewEngine()

	ack, err := engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "c-1",
		UserID:        "buyer1",
		Symbol:        "BTC-USD",
		Side:          SideBuy,
		Type:          OrderTypeLimit,
		Price:         100,
		Qty:           5,
	})
	if err != nil {
		t.Fatalf("place order failed: %v", err)
	}

	walletBefore := engine.Wallet("buyer1")
	if walletBefore.Reserved["USD"] != 500 {
		t.Fatalf("expected 500 USD reserved, got %d", walletBefore.Reserved["USD"])
	}

	cancelAck, err := engine.CancelOrder("buyer1", ack.OrderID)
	if err != nil {
		t.Fatalf("cancel order failed: %v", err)
	}
	if cancelAck.Status != OrderStatusCanceled {
		t.Fatalf("expected %s status, got %s", OrderStatusCanceled, cancelAck.Status)
	}

	walletAfter := engine.Wallet("buyer1")
	if walletAfter.Reserved["USD"] != 0 {
		t.Fatalf("expected 0 USD reserved after cancel, got %d", walletAfter.Reserved["USD"])
	}
	if walletAfter.Available["USD"] != 100000 {
		t.Fatalf("expected 100000 USD available after cancel, got %d", walletAfter.Available["USD"])
	}

	open := engine.OpenOrders("buyer1")
	if len(open) != 0 {
		t.Fatalf("expected 0 open orders after cancel, got %d", len(open))
	}
}

func TestCancelOrderReturnsErrorWhenOrderMissing(t *testing.T) {
	engine := NewEngine()

	_, err := engine.CancelOrder("buyer1", "ord-does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing order")
	}
}
