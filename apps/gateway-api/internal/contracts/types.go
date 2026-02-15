package contracts

import "time"

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

type OrderStatus string

const (
	OrderStatusAccepted      OrderStatus = "ACCEPTED"
	OrderStatusPartiallyFill OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled        OrderStatus = "FILLED"
	OrderStatusCanceled      OrderStatus = "CANCELED"
	OrderStatusRejected      OrderStatus = "REJECTED"
)

type PlaceOrderRequest struct {
	ClientOrderID string    `json:"clientOrderId"`
	UserID        string    `json:"userId"`
	Symbol        string    `json:"symbol"`
	Side          Side      `json:"side"`
	Type          OrderType `json:"type"`
	Price         int64     `json:"price,omitempty"`
	Qty           int64     `json:"qty"`
}

type OrderAck struct {
	OrderID       string      `json:"orderId"`
	Status        OrderStatus `json:"status"`
	FilledQty     int64       `json:"filledQty"`
	RemainingQty  int64       `json:"remainingQty"`
	AvgPrice      int64       `json:"avgPrice"`
	ClientOrderID string      `json:"clientOrderId,omitempty"`
	Symbol        string      `json:"symbol,omitempty"`
	TS            time.Time   `json:"ts"`
}

type Order struct {
	OrderID       string    `json:"orderId"`
	ClientOrderID string    `json:"clientOrderId"`
	UserID        string    `json:"userId"`
	Symbol        string    `json:"symbol"`
	Side          Side      `json:"side"`
	Type          OrderType `json:"type"`
	Price         int64     `json:"price"`
	Qty           int64     `json:"qty"`
	RemainingQty  int64     `json:"remainingQty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Wallet struct {
	UserID    string           `json:"userId"`
	Available map[string]int64 `json:"available"`
	Reserved  map[string]int64 `json:"reserved"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

type Execution struct {
	TradeID      string    `json:"tradeId"`
	Symbol       string    `json:"symbol"`
	Price        int64     `json:"price"`
	Qty          int64     `json:"qty"`
	MakerOrderID string    `json:"makerOrderId"`
	MakerUserID  string    `json:"makerUserId"`
	TakerOrderID string    `json:"takerOrderId"`
	TakerUserID  string    `json:"takerUserId"`
	TS           time.Time `json:"ts"`
}

type BookLevel struct {
	Price  int64 `json:"price"`
	Qty    int64 `json:"qty"`
	Orders int   `json:"orders"`
}

type OrderBookSnapshot struct {
	Symbol string      `json:"symbol"`
	Bids   []BookLevel `json:"bids"`
	Asks   []BookLevel `json:"asks"`
	TS     time.Time   `json:"ts"`
}

type Candle struct {
	Symbol      string    `json:"symbol"`
	Timeframe   string    `json:"timeframe"`
	BucketStart time.Time `json:"bucketStart"`
	Open        float64   `json:"open"`
	High        float64   `json:"high"`
	Low         float64   `json:"low"`
	Close       float64   `json:"close"`
	Volume      float64   `json:"volume"`
}

type ChartRenderRequest struct {
	Symbol    string `json:"symbol"`
	Timeframe string `json:"timeframe"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Theme     string `json:"theme,omitempty"`
}

type ChartRenderResponse struct {
	Cached       bool               `json:"cached"`
	CacheKey     string             `json:"cacheKey"`
	RenderID     string             `json:"renderId"`
	ArtifactType string             `json:"artifactType"`
	Artifact     string             `json:"artifact"`
	Meta         ChartRenderRequest `json:"meta"`
}
