package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Controller interface {
	Start() error
	Stop() error
	SetVolatility(volatility float64) error
	PauseSymbol(symbol string) error
	ResumeSymbol(symbol string) error
	EnsureSymbol(symbol string) error
	Running() bool
}

type Server struct {
	controller Controller
	mux        *http.ServeMux
}

func NewServer(controller Controller) *Server {
	s := &Server{controller: controller, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/v1/admin/sim/start", s.handleStart)
	s.mux.HandleFunc("/v1/admin/sim/stop", s.handleStop)
	s.mux.HandleFunc("/v1/admin/sim/volatility-profile", s.handleVolatility)
	s.mux.HandleFunc("/v1/admin/symbols/", s.handleSymbols)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.controller.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"running": s.controller.Running()})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.controller.Stop(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"running": s.controller.Running()})
}

func (s *Server) handleVolatility(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Volatility float64 `json:"volatility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := s.controller.SetVolatility(req.Volatility); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"running": s.controller.Running()})
}

func (s *Server) handleSymbols(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/admin/symbols/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	symbol := strings.TrimSpace(parts[0])
	action := strings.TrimSpace(parts[1])
	if symbol == "" {
		http.Error(w, "symbol is required", http.StatusBadRequest)
		return
	}

	var err error
	paused := false
	ensured := false
	switch action {
	case "pause":
		paused = true
		err = s.controller.PauseSymbol(symbol)
	case "resume":
		err = s.controller.ResumeSymbol(symbol)
	case "ensure":
		ensured = true
		err = s.controller.EnsureSymbol(symbol)
	default:
		http.NotFound(w, r)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"running": s.controller.Running(),
		"symbol":  symbol,
		"paused":  paused,
		"ensured": ensured,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
