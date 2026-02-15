package ledger

import (
	"context"
	"testing"
	"time"
)

type recordingSink struct {
	rows []ExecutionEvent
}

func (r *recordingSink) WriteExecution(_ context.Context, event ExecutionEvent) error {
	r.rows = append(r.rows, event)
	return nil
}

func TestServiceHandleWritesExecutionEvent(t *testing.T) {
	sink := &recordingSink{}
	svc := NewService(sink)

	event := ExecutionEvent{
		TradeID:    "trd-1",
		Symbol:     "BTC-USD",
		BuyUserID:  "buyer1",
		SellUserID: "seller1",
		Price:      101.25,
		Qty:        2.5,
		ExecutedAt: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
	}
	if err := svc.Handle(context.Background(), event); err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	if len(sink.rows) != 1 {
		t.Fatalf("expected 1 write, got %d", len(sink.rows))
	}
	if sink.rows[0].TradeID != "trd-1" {
		t.Fatalf("expected trade id trd-1, got %s", sink.rows[0].TradeID)
	}
}

func TestServiceHandleRejectsMissingTradeID(t *testing.T) {
	svc := NewService(&recordingSink{})
	err := svc.Handle(context.Background(), ExecutionEvent{Symbol: "BTC-USD", Price: 1, Qty: 1, ExecutedAt: time.Now().UTC()})
	if err == nil {
		t.Fatal("expected error for missing trade id")
	}
}
