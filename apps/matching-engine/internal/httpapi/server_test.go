package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kalency/apps/matching-engine/internal/matching"
)

func TestCreateOrderEndpoint(t *testing.T) {
	engine := matching.NewEngine()
	server := NewServer(engine)

	req := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{
		"clientOrderId":"c-1",
		"userId":"u1",
		"symbol":"BTC-USD",
		"side":"BUY",
		"type":"LIMIT",
		"price":100,
		"qty":10
	}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var ack matching.OrderAck
	if err := json.Unmarshal(rr.Body.Bytes(), &ack); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if ack.Status != matching.OrderStatusAccepted {
		t.Fatalf("expected status %s, got %s", matching.OrderStatusAccepted, ack.Status)
	}
	if ack.OrderID == "" {
		t.Fatal("expected non-empty order id")
	}
}

func TestListOpenOrdersEndpoint(t *testing.T) {
	engine := matching.NewEngine()
	server := NewServer(engine)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{
		"clientOrderId":"c-1",
		"userId":"u1",
		"symbol":"BTC-USD",
		"side":"BUY",
		"type":"LIMIT",
		"price":100,
		"qty":10
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	server.ServeHTTP(createRR, createReq)

	listReq := httptest.NewRequest(http.MethodGet, "/v1/orders/open/u1", nil)
	listRR := httptest.NewRecorder()
	server.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", listRR.Code)
	}

	var orders []matching.Order
	if err := json.Unmarshal(listRR.Body.Bytes(), &orders); err != nil {
		t.Fatalf("failed to decode orders response: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 open order, got %d", len(orders))
	}
	if orders[0].UserID != "u1" {
		t.Fatalf("expected user u1, got %s", orders[0].UserID)
	}
}

func TestListTradesEndpoint(t *testing.T) {
	engine := matching.NewEngine()
	engine.FundWallet("seller1", "BTC", 5)
	server := NewServer(engine)

	sellReq := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{
		"clientOrderId":"s-1",
		"userId":"seller1",
		"symbol":"BTC-USD",
		"side":"SELL",
		"type":"LIMIT",
		"price":100,
		"qty":5
	}`))
	sellReq.Header.Set("Content-Type", "application/json")
	sellRR := httptest.NewRecorder()
	server.ServeHTTP(sellRR, sellReq)
	if sellRR.Code != http.StatusCreated {
		t.Fatalf("expected sell status 201, got %d", sellRR.Code)
	}

	buyReq := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{
		"clientOrderId":"b-1",
		"userId":"buyer1",
		"symbol":"BTC-USD",
		"side":"BUY",
		"type":"MARKET",
		"qty":2
	}`))
	buyReq.Header.Set("Content-Type", "application/json")
	buyRR := httptest.NewRecorder()
	server.ServeHTTP(buyRR, buyReq)
	if buyRR.Code != http.StatusCreated {
		t.Fatalf("expected buy status 201, got %d", buyRR.Code)
	}

	tradesReq := httptest.NewRequest(http.MethodGet, "/v1/markets/BTC-USD/trades", nil)
	tradesRR := httptest.NewRecorder()
	server.ServeHTTP(tradesRR, tradesReq)
	if tradesRR.Code != http.StatusOK {
		t.Fatalf("expected trades status 200, got %d", tradesRR.Code)
	}

	var trades []matching.Execution
	if err := json.Unmarshal(tradesRR.Body.Bytes(), &trades); err != nil {
		t.Fatalf("failed to decode trades response: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Symbol != "BTC-USD" {
		t.Fatalf("expected BTC-USD symbol, got %s", trades[0].Symbol)
	}
	if trades[0].Qty != 2 {
		t.Fatalf("expected qty 2, got %d", trades[0].Qty)
	}
}

func TestOrderBookEndpoint(t *testing.T) {
	engine := matching.NewEngine()
	engine.FundWallet("seller1", "BTC", 5)
	engine.FundWallet("seller2", "BTC", 5)
	server := NewServer(engine)

	sellReq1 := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{
		"clientOrderId":"s-1",
		"userId":"seller1",
		"symbol":"BTC-USD",
		"side":"SELL",
		"type":"LIMIT",
		"price":110,
		"qty":2
	}`))
	sellReq1.Header.Set("Content-Type", "application/json")
	sellRR1 := httptest.NewRecorder()
	server.ServeHTTP(sellRR1, sellReq1)
	if sellRR1.Code != http.StatusCreated {
		t.Fatalf("expected first sell status 201, got %d", sellRR1.Code)
	}

	sellReq2 := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{
		"clientOrderId":"s-2",
		"userId":"seller2",
		"symbol":"BTC-USD",
		"side":"SELL",
		"type":"LIMIT",
		"price":111,
		"qty":3
	}`))
	sellReq2.Header.Set("Content-Type", "application/json")
	sellRR2 := httptest.NewRecorder()
	server.ServeHTTP(sellRR2, sellReq2)
	if sellRR2.Code != http.StatusCreated {
		t.Fatalf("expected second sell status 201, got %d", sellRR2.Code)
	}

	buyReq := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{
		"clientOrderId":"b-1",
		"userId":"buyer1",
		"symbol":"BTC-USD",
		"side":"BUY",
		"type":"LIMIT",
		"price":100,
		"qty":4
	}`))
	buyReq.Header.Set("Content-Type", "application/json")
	buyRR := httptest.NewRecorder()
	server.ServeHTTP(buyRR, buyReq)
	if buyRR.Code != http.StatusCreated {
		t.Fatalf("expected buy status 201, got %d", buyRR.Code)
	}

	bookReq := httptest.NewRequest(http.MethodGet, "/v1/markets/BTC-USD/book?depth=1", nil)
	bookRR := httptest.NewRecorder()
	server.ServeHTTP(bookRR, bookReq)
	if bookRR.Code != http.StatusOK {
		t.Fatalf("expected order book status 200, got %d", bookRR.Code)
	}

	var snapshot struct {
		Symbol string `json:"symbol"`
		Bids   []struct {
			Price int64 `json:"price"`
			Qty   int64 `json:"qty"`
		} `json:"bids"`
		Asks []struct {
			Price int64 `json:"price"`
			Qty   int64 `json:"qty"`
		} `json:"asks"`
	}
	if err := json.Unmarshal(bookRR.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode book response failed: %v", err)
	}
	if snapshot.Symbol != "BTC-USD" {
		t.Fatalf("expected symbol BTC-USD, got %s", snapshot.Symbol)
	}
	if len(snapshot.Bids) != 1 {
		t.Fatalf("expected 1 bid level, got %d", len(snapshot.Bids))
	}
	if len(snapshot.Asks) != 1 {
		t.Fatalf("expected 1 ask level, got %d", len(snapshot.Asks))
	}
	if snapshot.Bids[0].Price != 100 || snapshot.Bids[0].Qty != 4 {
		t.Fatalf("unexpected top bid %+v", snapshot.Bids[0])
	}
	if snapshot.Asks[0].Price != 110 || snapshot.Asks[0].Qty != 2 {
		t.Fatalf("unexpected top ask %+v", snapshot.Asks[0])
	}
}

func TestFundWalletEndpoint(t *testing.T) {
	engine := matching.NewEngine()
	server := NewServer(engine)

	fundReq := httptest.NewRequest(http.MethodPost, "/v1/admin/wallets/fund", strings.NewReader(`{
		"userId":"maker-bot",
		"asset":"BTC",
		"amount":50
	}`))
	fundReq.Header.Set("Content-Type", "application/json")
	fundRR := httptest.NewRecorder()
	server.ServeHTTP(fundRR, fundReq)

	if fundRR.Code != http.StatusOK {
		t.Fatalf("expected fund status 200, got %d", fundRR.Code)
	}

	walletReq := httptest.NewRequest(http.MethodGet, "/v1/wallet/maker-bot", nil)
	walletRR := httptest.NewRecorder()
	server.ServeHTTP(walletRR, walletReq)
	if walletRR.Code != http.StatusOK {
		t.Fatalf("expected wallet status 200, got %d", walletRR.Code)
	}

	var wallet matching.Wallet
	if err := json.Unmarshal(walletRR.Body.Bytes(), &wallet); err != nil {
		t.Fatalf("failed to decode wallet response: %v", err)
	}
	if wallet.Available["BTC"] != 50 {
		t.Fatalf("expected BTC available 50, got %d", wallet.Available["BTC"])
	}
}

func TestCORSPreflightForOrders(t *testing.T) {
	engine := matching.NewEngine()
	server := NewServer(engine)

	req := httptest.NewRequest(http.MethodOptions, "/v1/orders", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin=*, got %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
}
