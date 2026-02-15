package matching

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

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

const (
	defaultQuoteAsset   = "USD"
	defaultQuoteBalance = int64(100000)
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
	seq           int64

	BaseAsset        string `json:"-"`
	QuoteAsset       string `json:"-"`
	ReservedBaseQty  int64  `json:"-"`
	ReservedQuoteQty int64  `json:"-"`
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

type OpenOrdersStore interface {
	SetUserOrders(ctx context.Context, userID string, orders []Order) error
	GetUserOrders(ctx context.Context, userID string) ([]Order, bool, error)
}

type ExecutionSink interface {
	PublishExecution(ctx context.Context, execution Execution) error
}

type orderBook struct {
	bids []*Order
	asks []*Order
}

type Engine struct {
	mu              sync.Mutex
	books           map[string]*orderBook
	ordersByUser    map[string]map[string]*Order
	executions      map[string][]Execution
	wallets         map[string]*Wallet
	openOrdersStore OpenOrdersStore
	executionSink   ExecutionSink
	orderSeq        int64
	tradeSeq        int64
}

func NewEngine() *Engine {
	return NewEngineWithStoreAndSink(nil, nil)
}

func NewEngineWithStore(openOrdersStore OpenOrdersStore) *Engine {
	return NewEngineWithStoreAndSink(openOrdersStore, nil)
}

func NewEngineWithStoreAndSink(openOrdersStore OpenOrdersStore, executionSink ExecutionSink) *Engine {
	return &Engine{
		books:           make(map[string]*orderBook),
		ordersByUser:    make(map[string]map[string]*Order),
		executions:      make(map[string][]Execution),
		wallets:         make(map[string]*Wallet),
		openOrdersStore: openOrdersStore,
		executionSink:   executionSink,
	}
}

func (e *Engine) PlaceOrder(req PlaceOrderRequest) (OrderAck, error) {
	e.mu.Lock()

	if err := validate(req); err != nil {
		e.mu.Unlock()
		return OrderAck{}, err
	}

	baseAsset, quoteAsset, _ := parseSymbol(req.Symbol)

	e.orderSeq++
	order := &Order{
		OrderID:       fmt.Sprintf("ord-%d", e.orderSeq),
		ClientOrderID: req.ClientOrderID,
		UserID:        req.UserID,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Qty:           req.Qty,
		RemainingQty:  req.Qty,
		CreatedAt:     time.Now().UTC(),
		seq:           e.orderSeq,
		BaseAsset:     baseAsset,
		QuoteAsset:    quoteAsset,
	}

	book := e.ensureBook(req.Symbol)
	if err := e.reserveForOrderLocked(order, book); err != nil {
		e.mu.Unlock()
		return OrderAck{}, err
	}

	filled, avgPrice, touchedUsers, matchedExecutions, err := e.match(book, order)
	if err != nil {
		e.releaseOrderReservationLocked(order)
		e.mu.Unlock()
		return OrderAck{}, err
	}
	touchedUsers[order.UserID] = struct{}{}

	if order.Type == OrderTypeLimit && order.RemainingQty > 0 {
		e.addToBook(book, order)
		e.trackOpenOrder(order)
	}

	if order.Type == OrderTypeMarket && filled == 0 {
		e.releaseOrderReservationLocked(order)
		e.mu.Unlock()
		return OrderAck{}, errors.New("no liquidity for market order")
	}

	if order.Type == OrderTypeMarket || order.RemainingQty == 0 {
		e.releaseOrderReservationLocked(order)
	}

	status := OrderStatusAccepted
	switch {
	case filled > 0 && order.RemainingQty == 0:
		status = OrderStatusFilled
	case filled > 0 && order.RemainingQty > 0:
		status = OrderStatusPartiallyFill
	}

	ack := OrderAck{
		OrderID:       order.OrderID,
		Status:        status,
		FilledQty:     filled,
		RemainingQty:  order.RemainingQty,
		AvgPrice:      avgPrice,
		ClientOrderID: order.ClientOrderID,
		Symbol:        order.Symbol,
		TS:            time.Now().UTC(),
	}

	snapshots := make(map[string][]Order)
	if e.openOrdersStore != nil {
		for userID := range touchedUsers {
			snapshots[userID] = e.openOrdersSnapshotLocked(userID)
		}
	}

	e.mu.Unlock()

	if e.openOrdersStore != nil {
		ctx := context.Background()
		for userID, orders := range snapshots {
			_ = e.openOrdersStore.SetUserOrders(ctx, userID, orders)
		}
	}

	if e.executionSink != nil {
		ctx := context.Background()
		for _, execution := range matchedExecutions {
			_ = e.executionSink.PublishExecution(ctx, execution)
		}
	}

	return ack, nil
}

func (e *Engine) CancelOrder(userID, orderID string) (OrderAck, error) {
	e.mu.Lock()

	byUser, ok := e.ordersByUser[userID]
	if !ok {
		e.mu.Unlock()
		return OrderAck{}, errors.New("order not found")
	}

	order, ok := byUser[orderID]
	if !ok {
		e.mu.Unlock()
		return OrderAck{}, errors.New("order not found")
	}

	book := e.books[order.Symbol]
	if book == nil {
		e.mu.Unlock()
		return OrderAck{}, errors.New("order book not found")
	}

	remainingBeforeCancel := order.RemainingQty
	filledQty := order.Qty - remainingBeforeCancel

	e.removeFromBook(book, order)
	e.removeOpenOrder(order)
	e.releaseOrderReservationLocked(order)
	order.RemainingQty = 0

	ack := OrderAck{
		OrderID:       order.OrderID,
		Status:        OrderStatusCanceled,
		FilledQty:     filledQty,
		RemainingQty:  0,
		AvgPrice:      0,
		ClientOrderID: order.ClientOrderID,
		Symbol:        order.Symbol,
		TS:            time.Now().UTC(),
	}

	snapshot := []Order{}
	if e.openOrdersStore != nil {
		snapshot = e.openOrdersSnapshotLocked(userID)
	}

	e.mu.Unlock()

	if e.openOrdersStore != nil {
		_ = e.openOrdersStore.SetUserOrders(context.Background(), userID, snapshot)
	}

	return ack, nil
}

func (e *Engine) OpenOrders(userID string) []Order {
	if e.openOrdersStore != nil {
		orders, found, err := e.openOrdersStore.GetUserOrders(context.Background(), userID)
		if err == nil && found {
			sortByCreatedAt(orders)
			return orders
		}
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	return e.openOrdersSnapshotLocked(userID)
}

func (e *Engine) Executions(symbol string) []Execution {
	e.mu.Lock()
	defer e.mu.Unlock()

	entries := e.executions[symbol]
	out := make([]Execution, len(entries))
	copy(out, entries)
	return out
}

func (e *Engine) ListExecutions(symbol string, limit int) ([]Execution, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entries := e.executions[symbol]
	if len(entries) == 0 {
		return []Execution{}, nil
	}

	if limit <= 0 || limit >= len(entries) {
		out := make([]Execution, len(entries))
		copy(out, entries)
		return out, nil
	}

	start := len(entries) - limit
	out := make([]Execution, len(entries[start:]))
	copy(out, entries[start:])
	return out, nil
}

func (e *Engine) OrderBookSnapshot(symbol string, depth int) OrderBookSnapshot {
	if depth <= 0 {
		depth = 20
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	snapshot := OrderBookSnapshot{
		Symbol: symbol,
		Bids:   []BookLevel{},
		Asks:   []BookLevel{},
		TS:     time.Now().UTC(),
	}

	book, ok := e.books[symbol]
	if !ok || book == nil {
		return snapshot
	}

	snapshot.Bids = aggregateBookLevels(book.bids, depth)
	snapshot.Asks = aggregateBookLevels(book.asks, depth)
	return snapshot
}

func (e *Engine) FundWallet(userID, asset string, amount int64) {
	if amount <= 0 {
		return
	}
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	wallet := e.ensureWalletLocked(userID)
	wallet.Available[asset] += amount
	wallet.UpdatedAt = time.Now().UTC()
}

func (e *Engine) Wallet(userID string) Wallet {
	e.mu.Lock()
	defer e.mu.Unlock()

	wallet := e.ensureWalletLocked(userID)
	return copyWallet(wallet)
}

func validate(req PlaceOrderRequest) error {
	if req.UserID == "" {
		return errors.New("userId is required")
	}
	if req.Symbol == "" {
		return errors.New("symbol is required")
	}
	if _, _, err := parseSymbol(req.Symbol); err != nil {
		return err
	}
	if req.Qty <= 0 {
		return errors.New("qty must be positive")
	}
	if req.Side != SideBuy && req.Side != SideSell {
		return errors.New("side must be BUY or SELL")
	}
	if req.Type != OrderTypeLimit && req.Type != OrderTypeMarket {
		return errors.New("type must be LIMIT or MARKET")
	}
	if req.Type == OrderTypeLimit && req.Price <= 0 {
		return errors.New("price must be positive for LIMIT order")
	}
	return nil
}

func parseSymbol(symbol string) (string, string, error) {
	parts := strings.Split(symbol, "-")
	if len(parts) != 2 {
		return "", "", errors.New("symbol must be BASE-QUOTE format")
	}
	base := strings.ToUpper(strings.TrimSpace(parts[0]))
	quote := strings.ToUpper(strings.TrimSpace(parts[1]))
	if base == "" || quote == "" {
		return "", "", errors.New("symbol must be BASE-QUOTE format")
	}
	return base, quote, nil
}

func (e *Engine) ensureBook(symbol string) *orderBook {
	book, ok := e.books[symbol]
	if !ok {
		book = &orderBook{}
		e.books[symbol] = book
	}
	return book
}

func (e *Engine) trackOpenOrder(order *Order) {
	if _, ok := e.ordersByUser[order.UserID]; !ok {
		e.ordersByUser[order.UserID] = make(map[string]*Order)
	}
	e.ordersByUser[order.UserID][order.OrderID] = order
}

func (e *Engine) removeOpenOrder(order *Order) {
	if byUser, ok := e.ordersByUser[order.UserID]; ok {
		delete(byUser, order.OrderID)
		if len(byUser) == 0 {
			delete(e.ordersByUser, order.UserID)
		}
	}
}

func (e *Engine) match(book *orderBook, taker *Order) (filledQty int64, avgPrice int64, touchedUsers map[string]struct{}, matchedExecutions []Execution, err error) {
	touchedUsers = make(map[string]struct{})
	var weightedNotional int64

	for taker.RemainingQty > 0 {
		maker := e.bestMatch(book, taker)
		if maker == nil {
			break
		}

		tradeQty := minInt64(taker.RemainingQty, maker.RemainingQty)
		tradePrice := maker.Price

		if err := e.settleTradeLocked(taker, maker, tradeQty, tradePrice); err != nil {
			return 0, 0, nil, nil, err
		}

		taker.RemainingQty -= tradeQty
		maker.RemainingQty -= tradeQty
		filledQty += tradeQty
		weightedNotional += tradeQty * tradePrice
		touchedUsers[maker.UserID] = struct{}{}

		e.tradeSeq++
		execution := Execution{
			TradeID:      fmt.Sprintf("trd-%d", e.tradeSeq),
			Symbol:       taker.Symbol,
			Price:        tradePrice,
			Qty:          tradeQty,
			MakerOrderID: maker.OrderID,
			MakerUserID:  maker.UserID,
			TakerOrderID: taker.OrderID,
			TakerUserID:  taker.UserID,
			TS:           time.Now().UTC(),
		}
		e.executions[taker.Symbol] = append(e.executions[taker.Symbol], execution)
		matchedExecutions = append(matchedExecutions, execution)

		if maker.RemainingQty == 0 {
			e.removeFromBook(book, maker)
			e.removeOpenOrder(maker)
		}
	}

	if filledQty > 0 {
		avgPrice = weightedNotional / filledQty
	}
	return filledQty, avgPrice, touchedUsers, matchedExecutions, nil
}

func (e *Engine) settleTradeLocked(taker *Order, maker *Order, tradeQty int64, tradePrice int64) error {
	var buyer *Order
	var seller *Order
	if taker.Side == SideBuy {
		buyer = taker
		seller = maker
	} else {
		buyer = maker
		seller = taker
	}

	notional := tradeQty * tradePrice
	baseAsset := buyer.BaseAsset
	quoteAsset := buyer.QuoteAsset

	buyerWallet := e.ensureWalletLocked(buyer.UserID)
	sellerWallet := e.ensureWalletLocked(seller.UserID)

	if buyer.ReservedQuoteQty > 0 {
		reserveRelease := notional
		if buyer.Type == OrderTypeLimit {
			reserveRelease = buyer.Price * tradeQty
		}
		if reserveRelease > buyer.ReservedQuoteQty {
			reserveRelease = buyer.ReservedQuoteQty
		}
		if buyerWallet.Reserved[quoteAsset] < reserveRelease {
			return errors.New("buyer reserved quote balance underflow")
		}

		buyerWallet.Reserved[quoteAsset] -= reserveRelease
		buyer.ReservedQuoteQty -= reserveRelease

		switch {
		case reserveRelease > notional:
			buyerWallet.Available[quoteAsset] += reserveRelease - notional
		case reserveRelease < notional:
			extra := notional - reserveRelease
			if buyerWallet.Available[quoteAsset] < extra {
				return errors.New("insufficient quote balance")
			}
			buyerWallet.Available[quoteAsset] -= extra
		}
	} else {
		if buyerWallet.Available[quoteAsset] < notional {
			return errors.New("insufficient quote balance")
		}
		buyerWallet.Available[quoteAsset] -= notional
	}
	buyerWallet.Available[baseAsset] += tradeQty
	buyerWallet.UpdatedAt = time.Now().UTC()

	if seller.ReservedBaseQty > 0 {
		release := minInt64(tradeQty, seller.ReservedBaseQty)
		if sellerWallet.Reserved[baseAsset] < release {
			return errors.New("seller reserved base balance underflow")
		}
		sellerWallet.Reserved[baseAsset] -= release
		seller.ReservedBaseQty -= release

		if release < tradeQty {
			shortfall := tradeQty - release
			if sellerWallet.Available[baseAsset] < shortfall {
				return errors.New("insufficient base balance")
			}
			sellerWallet.Available[baseAsset] -= shortfall
		}
	} else {
		if sellerWallet.Available[baseAsset] < tradeQty {
			return errors.New("insufficient base balance")
		}
		sellerWallet.Available[baseAsset] -= tradeQty
	}

	sellerWallet.Available[quoteAsset] += notional
	sellerWallet.UpdatedAt = time.Now().UTC()

	return nil
}

func (e *Engine) reserveForOrderLocked(order *Order, book *orderBook) error {
	wallet := e.ensureWalletLocked(order.UserID)
	baseAsset := order.BaseAsset
	quoteAsset := order.QuoteAsset

	if order.Side == SideSell {
		if wallet.Available[baseAsset] < order.Qty {
			return errors.New("insufficient base balance")
		}
		wallet.Available[baseAsset] -= order.Qty
		wallet.Reserved[baseAsset] += order.Qty
		wallet.UpdatedAt = time.Now().UTC()
		order.ReservedBaseQty = order.Qty
		return nil
	}

	if order.Type == OrderTypeLimit {
		required := order.Price * order.Qty
		if wallet.Available[quoteAsset] < required {
			return errors.New("insufficient quote balance")
		}
		wallet.Available[quoteAsset] -= required
		wallet.Reserved[quoteAsset] += required
		wallet.UpdatedAt = time.Now().UTC()
		order.ReservedQuoteQty = required
		return nil
	}

	required := estimateMarketBuyNotional(book, order.Qty)
	if required == 0 {
		return nil
	}
	if wallet.Available[quoteAsset] < required {
		return errors.New("insufficient quote balance")
	}
	wallet.Available[quoteAsset] -= required
	wallet.Reserved[quoteAsset] += required
	wallet.UpdatedAt = time.Now().UTC()
	order.ReservedQuoteQty = required
	return nil
}

func estimateMarketBuyNotional(book *orderBook, qty int64) int64 {
	remaining := qty
	var notional int64
	for _, ask := range book.asks {
		if remaining <= 0 {
			break
		}
		take := minInt64(remaining, ask.RemainingQty)
		notional += take * ask.Price
		remaining -= take
	}
	return notional
}

func (e *Engine) releaseOrderReservationLocked(order *Order) {
	wallet := e.ensureWalletLocked(order.UserID)

	if order.ReservedQuoteQty > 0 {
		release := minInt64(order.ReservedQuoteQty, wallet.Reserved[order.QuoteAsset])
		wallet.Reserved[order.QuoteAsset] -= release
		wallet.Available[order.QuoteAsset] += release
		order.ReservedQuoteQty -= release
	}

	if order.ReservedBaseQty > 0 {
		release := minInt64(order.ReservedBaseQty, wallet.Reserved[order.BaseAsset])
		wallet.Reserved[order.BaseAsset] -= release
		wallet.Available[order.BaseAsset] += release
		order.ReservedBaseQty -= release
	}
	wallet.UpdatedAt = time.Now().UTC()
}

func (e *Engine) ensureWalletLocked(userID string) *Wallet {
	wallet, ok := e.wallets[userID]
	if !ok {
		wallet = &Wallet{
			UserID:    userID,
			Available: map[string]int64{defaultQuoteAsset: defaultQuoteBalance},
			Reserved:  map[string]int64{},
			UpdatedAt: time.Now().UTC(),
		}
		e.wallets[userID] = wallet
	}
	return wallet
}

func (e *Engine) bestMatch(book *orderBook, taker *Order) *Order {
	if taker.Side == SideBuy {
		if len(book.asks) == 0 {
			return nil
		}
		bestAsk := book.asks[0]
		if taker.Type == OrderTypeLimit && taker.Price < bestAsk.Price {
			return nil
		}
		return bestAsk
	}

	if len(book.bids) == 0 {
		return nil
	}
	bestBid := book.bids[0]
	if taker.Type == OrderTypeLimit && taker.Price > bestBid.Price {
		return nil
	}
	return bestBid
}

func (e *Engine) addToBook(book *orderBook, order *Order) {
	if order.Side == SideBuy {
		book.bids = append(book.bids, order)
		sort.SliceStable(book.bids, func(i, j int) bool {
			if book.bids[i].Price == book.bids[j].Price {
				return book.bids[i].seq < book.bids[j].seq
			}
			return book.bids[i].Price > book.bids[j].Price
		})
		return
	}

	book.asks = append(book.asks, order)
	sort.SliceStable(book.asks, func(i, j int) bool {
		if book.asks[i].Price == book.asks[j].Price {
			return book.asks[i].seq < book.asks[j].seq
		}
		return book.asks[i].Price < book.asks[j].Price
	})
}

func (e *Engine) removeFromBook(book *orderBook, order *Order) {
	if order.Side == SideBuy {
		for i := range book.bids {
			if book.bids[i].OrderID == order.OrderID {
				book.bids = append(book.bids[:i], book.bids[i+1:]...)
				return
			}
		}
		return
	}
	for i := range book.asks {
		if book.asks[i].OrderID == order.OrderID {
			book.asks = append(book.asks[:i], book.asks[i+1:]...)
			return
		}
	}
}

func aggregateBookLevels(entries []*Order, depth int) []BookLevel {
	if len(entries) == 0 || depth <= 0 {
		return []BookLevel{}
	}

	levels := make([]BookLevel, 0, depth)
	for _, entry := range entries {
		if entry == nil || entry.RemainingQty <= 0 {
			continue
		}

		levelCount := len(levels)
		if levelCount == 0 || levels[levelCount-1].Price != entry.Price {
			if len(levels) == depth {
				break
			}
			levels = append(levels, BookLevel{
				Price:  entry.Price,
				Qty:    entry.RemainingQty,
				Orders: 1,
			})
			continue
		}

		levels[levelCount-1].Qty += entry.RemainingQty
		levels[levelCount-1].Orders++
	}
	return levels
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func sortByCreatedAt(list []Order) {
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})
}

func copyWallet(wallet *Wallet) Wallet {
	available := make(map[string]int64, len(wallet.Available))
	for asset, amount := range wallet.Available {
		available[asset] = amount
	}
	reserved := make(map[string]int64, len(wallet.Reserved))
	for asset, amount := range wallet.Reserved {
		reserved[asset] = amount
	}
	return Wallet{
		UserID:    wallet.UserID,
		Available: available,
		Reserved:  reserved,
		UpdatedAt: wallet.UpdatedAt,
	}
}

func (e *Engine) openOrdersSnapshotLocked(userID string) []Order {
	orders := e.ordersByUser[userID]
	if len(orders) == 0 {
		return []Order{}
	}

	list := make([]Order, 0, len(orders))
	for _, o := range orders {
		list = append(list, *o)
	}
	sortByCreatedAt(list)
	return list
}
