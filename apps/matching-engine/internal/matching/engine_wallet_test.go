package matching

import "testing"

func TestRejectsLimitBuyWhenInsufficientQuoteBalance(t *testing.T) {
	engine := NewEngine()

	_, err := engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "c-1",
		UserID:        "u1",
		Symbol:        "BTC-USD",
		Side:          SideBuy,
		Type:          OrderTypeLimit,
		Price:         20000,
		Qty:           10,
	})
	if err == nil {
		t.Fatal("expected insufficient quote balance error")
	}
}

func TestRejectsSellWhenInsufficientBaseBalance(t *testing.T) {
	engine := NewEngine()

	_, err := engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "c-1",
		UserID:        "u1",
		Symbol:        "BTC-USD",
		Side:          SideSell,
		Type:          OrderTypeLimit,
		Price:         100,
		Qty:           1,
	})
	if err == nil {
		t.Fatal("expected insufficient base balance error")
	}
}

func TestWalletBalancesUpdateAfterFilledTrade(t *testing.T) {
	engine := NewEngine()
	engine.FundWallet("seller1", "BTC", 5)

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
		t.Fatalf("seed sell order failed: %v", err)
	}

	_, err = engine.PlaceOrder(PlaceOrderRequest{
		ClientOrderID: "b-1",
		UserID:        "buyer1",
		Symbol:        "BTC-USD",
		Side:          SideBuy,
		Type:          OrderTypeMarket,
		Qty:           5,
	})
	if err != nil {
		t.Fatalf("market buy failed: %v", err)
	}

	buyerWallet := engine.Wallet("buyer1")
	if got := buyerWallet.Available["USD"]; got != 99500 {
		t.Fatalf("expected buyer USD 99500, got %d", got)
	}
	if got := buyerWallet.Available["BTC"]; got != 5 {
		t.Fatalf("expected buyer BTC 5, got %d", got)
	}

	sellerWallet := engine.Wallet("seller1")
	if got := sellerWallet.Available["USD"]; got != 100500 {
		t.Fatalf("expected seller USD 100500, got %d", got)
	}
	if got := sellerWallet.Available["BTC"]; got != 0 {
		t.Fatalf("expected seller BTC 0, got %d", got)
	}
}
