package ledger

import (
	"context"
	"errors"
	"strings"
	"time"
)

type ExecutionSink interface {
	WriteExecution(ctx context.Context, event ExecutionEvent) error
}

type Service struct {
	sink ExecutionSink
}

func NewService(sink ExecutionSink) *Service {
	return &Service{sink: sink}
}

func (s *Service) Handle(ctx context.Context, event ExecutionEvent) error {
	if s.sink == nil {
		return errors.New("execution sink is required")
	}
	event.TradeID = strings.TrimSpace(event.TradeID)
	event.Symbol = strings.TrimSpace(event.Symbol)
	event.BuyUserID = strings.TrimSpace(event.BuyUserID)
	event.SellUserID = strings.TrimSpace(event.SellUserID)

	if event.TradeID == "" {
		return errors.New("trade id is required")
	}
	if event.Symbol == "" {
		return errors.New("symbol is required")
	}
	if event.Price <= 0 {
		return errors.New("price must be positive")
	}
	if event.Qty <= 0 {
		return errors.New("qty must be positive")
	}
	if event.ExecutedAt.IsZero() {
		event.ExecutedAt = time.Now().UTC()
	}

	return s.sink.WriteExecution(ctx, event)
}
