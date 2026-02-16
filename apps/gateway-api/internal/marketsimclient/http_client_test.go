package marketsimclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPauseSymbolForwardsRequest(t *testing.T) {
	var calledPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"symbol":"BTC-USD","paused":true}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	_, err := client.PauseSymbol("BTC-USD")
	if err != nil {
		t.Fatalf("pause symbol failed: %v", err)
	}
	if calledPath != "/v1/admin/symbols/BTC-USD/pause" {
		t.Fatalf("unexpected path %q", calledPath)
	}
}

func TestSetVolatilityForwardsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/admin/sim/volatility-profile" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"running":true}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	_, err := client.SetVolatility(0.02)
	if err != nil {
		t.Fatalf("set volatility failed: %v", err)
	}
}

func TestEnsureSymbolForwardsRequest(t *testing.T) {
	var calledPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"symbol":"SOL-USD","ensured":true}`))
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	_, err := client.EnsureSymbol("SOL-USD")
	if err != nil {
		t.Fatalf("ensure symbol failed: %v", err)
	}
	if calledPath != "/v1/admin/symbols/SOL-USD/ensure" {
		t.Fatalf("unexpected path %q", calledPath)
	}
}
