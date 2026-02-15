package matching

import (
	"context"
	"testing"
)

type fakeExecutionSink struct {
	events []Execution
}

func (f *fakeExecutionSink) PublishExecution(_ context.Context, execution Execution) error {
	f.events = append(f.events, execution)
	return nil
}

func TestEnginePublishesExecutionsToSinkInMatchOrder(t *testing.T) {
	sink := &fakeExecutionSink{}
	engine := NewEngineWithStoreAndSink(nil, sink)
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

	if len(sink.events) != 2 {
		t.Fatalf("expected 2 published executions, got %d", len(sink.events))
	}
	if sink.events[0].MakerUserID != "seller1" || sink.events[0].Qty != 5 {
		t.Fatalf("expected first published event from seller1 qty=5, got maker=%s qty=%d", sink.events[0].MakerUserID, sink.events[0].Qty)
	}
	if sink.events[1].MakerUserID != "seller2" || sink.events[1].Qty != 2 {
		t.Fatalf("expected second published event from seller2 qty=2, got maker=%s qty=%d", sink.events[1].MakerUserID, sink.events[1].Qty)
	}
}
