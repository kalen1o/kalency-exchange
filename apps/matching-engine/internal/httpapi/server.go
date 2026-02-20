package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"kalency/apps/matching-engine/internal/matching"
)

type TradeSource interface {
	ListExecutions(symbol string, limit int) ([]matching.Execution, error)
}

type Server struct {
	engine      *matching.Engine
	tradeSource TradeSource
	mux         *http.ServeMux
}

func NewServer(engine *matching.Engine, tradeSource ...TradeSource) *Server {
	selectedTradeSource := TradeSource(engine)
	if len(tradeSource) > 0 && tradeSource[0] != nil {
		selectedTradeSource = tradeSource[0]
	}

	s := &Server{
		engine:      engine,
		tradeSource: selectedTradeSource,
		mux:         http.NewServeMux(),
	}
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
	s.mux.HandleFunc("/v1/orders", s.handleOrders)
	s.mux.HandleFunc("/v1/orders/", s.handleOrderByID)
	s.mux.HandleFunc("/v1/orders/open/", s.handleOpenOrders)
	s.mux.HandleFunc("/v1/wallet/", s.handleWallet)
	s.mux.HandleFunc("/v1/admin/wallets/fund", s.handleFundWallet)
	s.mux.HandleFunc("/v1/markets/", s.handleMarkets)
	s.mux.HandleFunc("/healthz", s.handleHealth)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req matching.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	ack, err := s.engine.PlaceOrder(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, ack)
}

func (s *Server) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/v1/orders/")
	if orderID == "" {
		http.Error(w, "order id is required", http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId query is required", http.StatusBadRequest)
		return
	}

	ack, err := s.engine.CancelOrder(userID, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, ack)
}

func (s *Server) handleOpenOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/v1/orders/open/")
	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	orders := s.engine.OpenOrders(userID)
	writeJSON(w, http.StatusOK, orders)
}

func (s *Server) handleWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/v1/wallet/")
	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	wallet := s.engine.Wallet(userID)
	writeJSON(w, http.StatusOK, wallet)
}

func (s *Server) handleFundWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID string `json:"userId"`
		Asset  string `json:"asset"`
		Amount int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	req.UserID = strings.TrimSpace(req.UserID)
	req.Asset = strings.TrimSpace(req.Asset)
	if req.UserID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}
	if req.Asset == "" {
		http.Error(w, "asset is required", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}

	s.engine.FundWallet(req.UserID, req.Asset, req.Amount)
	writeJSON(w, http.StatusOK, s.engine.Wallet(req.UserID))
}

func (s *Server) handleMarkets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/markets/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	symbol := parts[0]
	switch parts[1] {
	case "trades":
		limit := 100
		if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
			parsed, err := strconv.Atoi(rawLimit)
			if err != nil || parsed <= 0 {
				http.Error(w, "limit must be a positive integer", http.StatusBadRequest)
				return
			}
			limit = parsed
		}

		trades, err := s.tradeSource.ListExecutions(symbol, limit)
		if err != nil {
			http.Error(w, "failed to load trades", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, trades)
	case "book":
		depth := 20
		if rawDepth := r.URL.Query().Get("depth"); rawDepth != "" {
			parsed, err := strconv.Atoi(rawDepth)
			if err != nil || parsed <= 0 {
				http.Error(w, "depth must be a positive integer", http.StatusBadRequest)
				return
			}
			depth = parsed
		}

		snapshot := s.engine.OrderBookSnapshot(symbol, depth)
		writeJSON(w, http.StatusOK, snapshot)
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
}
