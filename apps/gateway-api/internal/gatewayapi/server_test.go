package gatewayapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"kalency/apps/gateway-api/internal/contracts"
)

type fakeTradingService struct {
	lastPlaceReq contracts.PlaceOrderRequest
	walletByUser map[string]contracts.Wallet
	bookBySymbol map[string]contracts.OrderBookSnapshot
}

func (f *fakeTradingService) PlaceOrder(req contracts.PlaceOrderRequest) (contracts.OrderAck, error) {
	f.lastPlaceReq = req
	return contracts.OrderAck{OrderID: "ord-1", Status: contracts.OrderStatusAccepted}, nil
}

func (f *fakeTradingService) CancelOrder(userID, orderID string) (contracts.OrderAck, error) {
	return contracts.OrderAck{OrderID: orderID, Status: contracts.OrderStatusCanceled}, nil
}

func (f *fakeTradingService) OpenOrders(userID string) ([]contracts.Order, error) {
	return []contracts.Order{}, nil
}

func (f *fakeTradingService) Wallet(userID string) (contracts.Wallet, error) {
	if wallet, ok := f.walletByUser[userID]; ok {
		return wallet, nil
	}
	return contracts.Wallet{UserID: userID, Available: map[string]int64{"USD": 100000}, Reserved: map[string]int64{}}, nil
}

func (f *fakeTradingService) ListExecutions(symbol string, limit int) ([]contracts.Execution, error) {
	return []contracts.Execution{}, nil
}

func (f *fakeTradingService) ListOrderBook(symbol string, depth int) (contracts.OrderBookSnapshot, error) {
	if snapshot, ok := f.bookBySymbol[symbol]; ok {
		return snapshot, nil
	}
	return contracts.OrderBookSnapshot{Symbol: symbol}, nil
}

func TestWalletEndpointRequiresAuth(t *testing.T) {
	svc := &fakeTradingService{walletByUser: map[string]contracts.Wallet{}}
	app := NewServer(Config{JWTSecret: "secret", APIKeys: map[string]string{}}, svc)

	req, _ := http.NewRequest(http.MethodGet, "/v1/wallet", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
}

func TestJWTTokenAndWalletFlow(t *testing.T) {
	svc := &fakeTradingService{walletByUser: map[string]contracts.Wallet{
		"u1": {UserID: "u1", Available: map[string]int64{"USD": 999}, Reserved: map[string]int64{}},
	}}
	app := NewServer(Config{JWTSecret: "secret", APIKeys: map[string]string{}}, svc)

	body, _ := json.Marshal(map[string]string{"userId": "u1"})
	tokenReq, _ := http.NewRequest(http.MethodPost, "/v1/auth/token", bytes.NewReader(body))
	tokenReq.Header.Set("Content-Type", "application/json")
	tokenRes, err := app.Test(tokenReq)
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	if tokenRes.StatusCode != http.StatusOK {
		t.Fatalf("expected token status 200, got %d", tokenRes.StatusCode)
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(tokenRes.Body).Decode(&tokenResponse); err != nil {
		t.Fatalf("decode token response failed: %v", err)
	}
	if tokenResponse.Token == "" {
		t.Fatal("expected non-empty token")
	}

	walletReq, _ := http.NewRequest(http.MethodGet, "/v1/wallet", nil)
	walletReq.Header.Set("Authorization", "Bearer "+tokenResponse.Token)
	walletRes, err := app.Test(walletReq)
	if err != nil {
		t.Fatalf("wallet request failed: %v", err)
	}
	if walletRes.StatusCode != http.StatusOK {
		t.Fatalf("expected wallet status 200, got %d", walletRes.StatusCode)
	}

	var wallet contracts.Wallet
	if err := json.NewDecoder(walletRes.Body).Decode(&wallet); err != nil {
		t.Fatalf("decode wallet failed: %v", err)
	}
	if wallet.UserID != "u1" {
		t.Fatalf("expected wallet user u1, got %s", wallet.UserID)
	}
}

func TestPlaceOrderUsesAPIKeyIdentity(t *testing.T) {
	svc := &fakeTradingService{walletByUser: map[string]contracts.Wallet{}}
	app := NewServer(Config{JWTSecret: "secret", APIKeys: map[string]string{"demo-key": "u1"}}, svc)

	body := []byte(`{"clientOrderId":"c-1","symbol":"BTC-USD","side":"BUY","type":"LIMIT","price":100,"qty":1}`)
	req, _ := http.NewRequest(http.MethodPost, "/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "demo-key")

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("place order request failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.StatusCode)
	}
	if svc.lastPlaceReq.UserID != "u1" {
		t.Fatalf("expected userId u1 from API key, got %s", svc.lastPlaceReq.UserID)
	}
}

type fakeCandleService struct {
	candles []contracts.Candle
}

func (f *fakeCandleService) ListCandles(symbol, timeframe string, from, to time.Time) ([]contracts.Candle, error) {
	return f.candles, nil
}

type fakeAdminService struct {
	startCalled       bool
	stopCalled        bool
	lastVolatility    float64
	lastPausedSymbol  string
	lastResumedSymbol string
	lastEnsuredSymbol string
}

func (f *fakeAdminService) StartSimulator() (map[string]any, error) {
	f.startCalled = true
	return map[string]any{"running": true}, nil
}

func (f *fakeAdminService) StopSimulator() (map[string]any, error) {
	f.stopCalled = true
	return map[string]any{"running": false}, nil
}

func (f *fakeAdminService) SetVolatility(volatility float64) (map[string]any, error) {
	f.lastVolatility = volatility
	return map[string]any{"running": true}, nil
}

func (f *fakeAdminService) PauseSymbol(symbol string) (map[string]any, error) {
	f.lastPausedSymbol = symbol
	return map[string]any{"symbol": symbol, "paused": true}, nil
}

func (f *fakeAdminService) ResumeSymbol(symbol string) (map[string]any, error) {
	f.lastResumedSymbol = symbol
	return map[string]any{"symbol": symbol, "paused": false}, nil
}

func (f *fakeAdminService) EnsureSymbol(symbol string) (map[string]any, error) {
	f.lastEnsuredSymbol = symbol
	return map[string]any{"symbol": symbol, "ensured": true}, nil
}

func TestCandlesEndpoint(t *testing.T) {
	tf := time.Date(2026, 2, 15, 0, 1, 0, 0, time.UTC)
	candleSvc := &fakeCandleService{candles: []contracts.Candle{{Symbol: "BTC-USD", Timeframe: "1m", BucketStart: tf, Open: 100, High: 101, Low: 99, Close: 100.5, Volume: 10}}}
	app := NewServer(Config{JWTSecret: "secret", APIKeys: map[string]string{"demo-key": "u1"}, CandleService: candleSvc}, &fakeTradingService{walletByUser: map[string]contracts.Wallet{}})

	req, _ := http.NewRequest(http.MethodGet, "/v1/markets/BTC-USD/candles?tf=1m&from=2026-02-15T00:00:00Z&to=2026-02-15T00:05:00Z", nil)
	req.Header.Set("X-API-Key", "demo-key")
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("candles request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var candles []contracts.Candle
	if err := json.NewDecoder(res.Body).Decode(&candles); err != nil {
		t.Fatalf("decode candles failed: %v", err)
	}
	if len(candles) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(candles))
	}
}

func TestOrderBookEndpoint(t *testing.T) {
	svc := &fakeTradingService{
		walletByUser: map[string]contracts.Wallet{},
		bookBySymbol: map[string]contracts.OrderBookSnapshot{
			"BTC-USD": {
				Symbol: "BTC-USD",
				Bids:   []contracts.BookLevel{{Price: 100, Qty: 3, Orders: 2}},
				Asks:   []contracts.BookLevel{{Price: 110, Qty: 5, Orders: 2}},
			},
		},
	}
	app := NewServer(Config{JWTSecret: "secret", APIKeys: map[string]string{"demo-key": "u1"}}, svc)

	req, _ := http.NewRequest(http.MethodGet, "/v1/markets/BTC-USD/book?depth=1", nil)
	req.Header.Set("X-API-Key", "demo-key")

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("order book request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	var snapshot contracts.OrderBookSnapshot
	if err := json.NewDecoder(res.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode order book failed: %v", err)
	}
	if snapshot.Symbol != "BTC-USD" {
		t.Fatalf("expected BTC-USD symbol, got %s", snapshot.Symbol)
	}
	if len(snapshot.Bids) != 1 || snapshot.Bids[0].Price != 100 {
		t.Fatalf("unexpected bids: %+v", snapshot.Bids)
	}
	if len(snapshot.Asks) != 1 || snapshot.Asks[0].Price != 110 {
		t.Fatalf("unexpected asks: %+v", snapshot.Asks)
	}
}

func TestAdminSimulatorEndpoints(t *testing.T) {
	adminSvc := &fakeAdminService{}
	app := NewServer(Config{
		JWTSecret:    "secret",
		APIKeys:      map[string]string{"demo-key": "u1"},
		AdminService: adminSvc,
	}, &fakeTradingService{walletByUser: map[string]contracts.Wallet{}})

	startReq, _ := http.NewRequest(http.MethodPost, "/v1/admin/sim/start", nil)
	startReq.Header.Set("X-API-Key", "demo-key")
	startRes, err := app.Test(startReq)
	if err != nil {
		t.Fatalf("admin start request failed: %v", err)
	}
	if startRes.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d", startRes.StatusCode)
	}
	if !adminSvc.startCalled {
		t.Fatal("expected admin start to call service")
	}

	volReq, _ := http.NewRequest(http.MethodPost, "/v1/admin/sim/volatility-profile", bytes.NewReader([]byte(`{"volatility":0.05}`)))
	volReq.Header.Set("Content-Type", "application/json")
	volReq.Header.Set("X-API-Key", "demo-key")
	volRes, err := app.Test(volReq)
	if err != nil {
		t.Fatalf("admin volatility request failed: %v", err)
	}
	if volRes.StatusCode != http.StatusOK {
		t.Fatalf("expected volatility 200, got %d", volRes.StatusCode)
	}
	if adminSvc.lastVolatility != 0.05 {
		t.Fatalf("expected volatility 0.05, got %f", adminSvc.lastVolatility)
	}
}

func TestAdminSymbolPauseResumeEndpoints(t *testing.T) {
	adminSvc := &fakeAdminService{}
	app := NewServer(Config{
		JWTSecret:    "secret",
		APIKeys:      map[string]string{"demo-key": "u1"},
		AdminService: adminSvc,
	}, &fakeTradingService{walletByUser: map[string]contracts.Wallet{}})

	pauseReq, _ := http.NewRequest(http.MethodPost, "/v1/admin/symbols/BTC-USD/pause", nil)
	pauseReq.Header.Set("X-API-Key", "demo-key")
	pauseRes, err := app.Test(pauseReq)
	if err != nil {
		t.Fatalf("pause request failed: %v", err)
	}
	if pauseRes.StatusCode != http.StatusOK {
		t.Fatalf("expected pause 200, got %d", pauseRes.StatusCode)
	}
	if adminSvc.lastPausedSymbol != "BTC-USD" {
		t.Fatalf("expected paused symbol BTC-USD, got %s", adminSvc.lastPausedSymbol)
	}

	resumeReq, _ := http.NewRequest(http.MethodPost, "/v1/admin/symbols/BTC-USD/resume", nil)
	resumeReq.Header.Set("X-API-Key", "demo-key")
	resumeRes, err := app.Test(resumeReq)
	if err != nil {
		t.Fatalf("resume request failed: %v", err)
	}
	if resumeRes.StatusCode != http.StatusOK {
		t.Fatalf("expected resume 200, got %d", resumeRes.StatusCode)
	}
	if adminSvc.lastResumedSymbol != "BTC-USD" {
		t.Fatalf("expected resumed symbol BTC-USD, got %s", adminSvc.lastResumedSymbol)
	}
}

func TestAdminSymbolEnsureEndpoint(t *testing.T) {
	adminSvc := &fakeAdminService{}
	app := NewServer(Config{
		JWTSecret:    "secret",
		APIKeys:      map[string]string{"demo-key": "u1"},
		AdminService: adminSvc,
	}, &fakeTradingService{walletByUser: map[string]contracts.Wallet{}})

	ensureReq, _ := http.NewRequest(http.MethodPost, "/v1/admin/symbols/SOL-USD/ensure", nil)
	ensureReq.Header.Set("X-API-Key", "demo-key")
	ensureRes, err := app.Test(ensureReq)
	if err != nil {
		t.Fatalf("ensure request failed: %v", err)
	}
	if ensureRes.StatusCode != http.StatusOK {
		t.Fatalf("expected ensure 200, got %d", ensureRes.StatusCode)
	}
	if adminSvc.lastEnsuredSymbol != "SOL-USD" {
		t.Fatalf("expected ensured symbol SOL-USD, got %s", adminSvc.lastEnsuredSymbol)
	}
}
