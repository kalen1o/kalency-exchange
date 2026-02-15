package chartclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kalency/apps/gateway-api/internal/contracts"
)

func TestHTTPClientRenderChart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/charts/render" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"cached":       false,
			"cacheKey":     "abc",
			"renderId":     "rid-1",
			"artifactType": "image/svg+xml",
			"artifact":     "<svg/>",
			"meta": map[string]any{
				"symbol":    "BTC-USD",
				"timeframe": "1m",
			},
		})
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	res, err := client.RenderChart(contracts.ChartRenderRequest{Symbol: "BTC-USD", Timeframe: "1m"})
	if err != nil {
		t.Fatalf("render chart failed: %v", err)
	}
	if res.RenderID != "rid-1" {
		t.Fatalf("expected render id rid-1, got %s", res.RenderID)
	}
}
