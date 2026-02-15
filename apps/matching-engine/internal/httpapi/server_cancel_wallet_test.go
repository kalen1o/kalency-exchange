package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kalency/apps/matching-engine/internal/matching"
)

func TestCancelOrderEndpoint(t *testing.T) {
	engine := matching.NewEngine()
	server := NewServer(engine)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader(`{
		"clientOrderId":"c-1",
		"userId":"u1",
		"symbol":"BTC-USD",
		"side":"BUY",
		"type":"LIMIT",
		"price":100,
		"qty":5
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	server.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRR.Code)
	}

	var placed matching.OrderAck
	if err := json.Unmarshal(createRR.Body.Bytes(), &placed); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	cancelReq := httptest.NewRequest(http.MethodDelete, "/v1/orders/"+placed.OrderID+"?userId=u1", nil)
	cancelRR := httptest.NewRecorder()
	server.ServeHTTP(cancelRR, cancelReq)
	if cancelRR.Code != http.StatusOK {
		t.Fatalf("expected cancel status 200, got %d", cancelRR.Code)
	}

	var canceled matching.OrderAck
	if err := json.Unmarshal(cancelRR.Body.Bytes(), &canceled); err != nil {
		t.Fatalf("failed to decode cancel response: %v", err)
	}
	if canceled.Status != matching.OrderStatusCanceled {
		t.Fatalf("expected status %s, got %s", matching.OrderStatusCanceled, canceled.Status)
	}
}

func TestWalletEndpoint(t *testing.T) {
	engine := matching.NewEngine()
	engine.FundWallet("u1", "BTC", 3)
	server := NewServer(engine)

	req := httptest.NewRequest(http.MethodGet, "/v1/wallet/u1", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var wallet matching.Wallet
	if err := json.Unmarshal(rr.Body.Bytes(), &wallet); err != nil {
		t.Fatalf("failed to decode wallet response: %v", err)
	}
	if wallet.UserID != "u1" {
		t.Fatalf("expected wallet user u1, got %s", wallet.UserID)
	}
	if wallet.Available["BTC"] != 3 {
		t.Fatalf("expected BTC balance 3, got %d", wallet.Available["BTC"])
	}
}
