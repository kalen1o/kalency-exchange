package store

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"kalency/apps/market-sim/internal/sim"
)

type capturedRequest struct {
	Path string
	Body map[string]any
}

func TestMatchingOrderSinkPublishesFundingAndOrders(t *testing.T) {
	var mu sync.Mutex
	requests := make([]capturedRequest, 0)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		mu.Lock()
		requests = append(requests, capturedRequest{Path: r.URL.Path, Body: body})
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	sink := NewMatchingOrderSink(srv.URL)
	tick := sim.Tick{Symbol: "BTC-USD", Price: 101.7, Volume: 1.9, TS: time.Now().UTC()}

	if err := sink.PublishTick(context.Background(), tick); err != nil {
		t.Fatalf("publish tick failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 6 {
		t.Fatalf("expected 6 requests (funding + maker order + taker order), got %d", len(requests))
	}
	if requests[0].Path != "/v1/admin/wallets/fund" {
		t.Fatalf("expected first request to fund endpoint, got %s", requests[0].Path)
	}
	if requests[4].Path != "/v1/orders" || requests[5].Path != "/v1/orders" {
		t.Fatalf("expected order requests to /v1/orders, got %s and %s", requests[4].Path, requests[5].Path)
	}
	if requests[4].Body["type"] != "LIMIT" || requests[5].Body["type"] != "MARKET" {
		t.Fatalf("expected LIMIT then MARKET order types, got %v then %v", requests[4].Body["type"], requests[5].Body["type"])
	}
}
