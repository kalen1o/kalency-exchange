package ledger

import "time"

type ExecutionEvent struct {
	TradeID    string
	Symbol     string
	BuyUserID  string
	SellUserID string
	Price      float64
	Qty        float64
	ExecutedAt time.Time
}
