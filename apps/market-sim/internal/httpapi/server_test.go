package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeController struct {
	running    bool
	volatility float64
	paused     map[string]bool
	ensured    map[string]bool
}

func (f *fakeController) Start() error {
	f.running = true
	return nil
}

func (f *fakeController) Stop() error {
	f.running = false
	return nil
}

func (f *fakeController) SetVolatility(volatility float64) error {
	f.volatility = volatility
	return nil
}

func (f *fakeController) Running() bool {
	return f.running
}

func (f *fakeController) PauseSymbol(symbol string) error {
	if f.paused == nil {
		f.paused = map[string]bool{}
	}
	f.paused[symbol] = true
	return nil
}

func (f *fakeController) ResumeSymbol(symbol string) error {
	if f.paused == nil {
		f.paused = map[string]bool{}
	}
	delete(f.paused, symbol)
	return nil
}

func (f *fakeController) EnsureSymbol(symbol string) error {
	if f.ensured == nil {
		f.ensured = map[string]bool{}
	}
	f.ensured[symbol] = true
	return nil
}

func TestHealthEndpoint(t *testing.T) {
	controller := &fakeController{}
	server := NewServer(controller)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != "ok" {
		t.Fatalf("expected body ok, got %q", rr.Body.String())
	}
}

func TestStartStopEndpoints(t *testing.T) {
	controller := &fakeController{}
	server := NewServer(controller)

	startReq := httptest.NewRequest(http.MethodPost, "/v1/admin/sim/start", nil)
	startRR := httptest.NewRecorder()
	server.ServeHTTP(startRR, startReq)
	if startRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 on start, got %d", startRR.Code)
	}
	if !controller.running {
		t.Fatal("expected simulator running after start")
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/v1/admin/sim/stop", nil)
	stopRR := httptest.NewRecorder()
	server.ServeHTTP(stopRR, stopReq)
	if stopRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 on stop, got %d", stopRR.Code)
	}
	if controller.running {
		t.Fatal("expected simulator stopped after stop")
	}
}

func TestVolatilityEndpoint(t *testing.T) {
	controller := &fakeController{}
	server := NewServer(controller)

	req := httptest.NewRequest(http.MethodPost, "/v1/admin/sim/volatility-profile", strings.NewReader(`{"volatility":0.02}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if controller.volatility != 0.02 {
		t.Fatalf("expected volatility 0.02, got %f", controller.volatility)
	}
}

func TestPauseResumeSymbolEndpoints(t *testing.T) {
	controller := &fakeController{}
	server := NewServer(controller)

	pauseReq := httptest.NewRequest(http.MethodPost, "/v1/admin/symbols/BTC-USD/pause", nil)
	pauseRR := httptest.NewRecorder()
	server.ServeHTTP(pauseRR, pauseReq)

	if pauseRR.Code != http.StatusOK {
		t.Fatalf("expected pause status 200, got %d", pauseRR.Code)
	}
	if !controller.paused["BTC-USD"] {
		t.Fatal("expected BTC-USD paused after pause endpoint")
	}

	resumeReq := httptest.NewRequest(http.MethodPost, "/v1/admin/symbols/BTC-USD/resume", nil)
	resumeRR := httptest.NewRecorder()
	server.ServeHTTP(resumeRR, resumeReq)

	if resumeRR.Code != http.StatusOK {
		t.Fatalf("expected resume status 200, got %d", resumeRR.Code)
	}
	if controller.paused["BTC-USD"] {
		t.Fatal("expected BTC-USD resumed after resume endpoint")
	}
}

func TestEnsureSymbolEndpoint(t *testing.T) {
	controller := &fakeController{}
	server := NewServer(controller)

	ensureReq := httptest.NewRequest(http.MethodPost, "/v1/admin/symbols/SOL-USD/ensure", nil)
	ensureRR := httptest.NewRecorder()
	server.ServeHTTP(ensureRR, ensureReq)

	if ensureRR.Code != http.StatusOK {
		t.Fatalf("expected ensure status 200, got %d", ensureRR.Code)
	}
	if !controller.ensured["SOL-USD"] {
		t.Fatal("expected SOL-USD ensured after ensure endpoint")
	}
}
