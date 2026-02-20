package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"kalency/apps/market-sim/internal/sim"
)

const initialBotFunding = int64(1_000_000)

type MatchingOrderSink struct {
	baseURL string
	client  *http.Client

	mu     sync.Mutex
	seq    int64
	funded map[string]struct{}
}

func NewMatchingOrderSink(baseURL string) *MatchingOrderSink {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &MatchingOrderSink{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
		funded:  map[string]struct{}{},
	}
}

func (s *MatchingOrderSink) PublishTick(ctx context.Context, tick sim.Tick) error {
	symbol := strings.ToUpper(strings.TrimSpace(tick.Symbol))
	baseAsset, quoteAsset, err := parseSymbol(symbol)
	if err != nil {
		return err
	}
	makerUserID := "sim-maker-" + symbol
	takerUserID := "sim-taker-" + symbol

	if err := s.ensureFunding(ctx, symbol, makerUserID, takerUserID, baseAsset, quoteAsset); err != nil {
		return err
	}

	price := int64(math.Round(tick.Price))
	if price < 1 {
		price = 1
	}
	qty := int64(math.Round(tick.Volume))
	if qty < 1 {
		qty = 1
	}

	makerOrder := orderPayload{
		ClientOrderID: s.nextOrderID(),
		UserID:        makerUserID,
		Symbol:        symbol,
		Qty:           qty,
		Type:          "LIMIT",
		Price:         price,
	}
	takerOrder := orderPayload{
		ClientOrderID: s.nextOrderID(),
		UserID:        takerUserID,
		Symbol:        symbol,
		Qty:           qty,
		Type:          "MARKET",
	}

	// Negative delta implies sell pressure; otherwise buy pressure.
	if tick.Delta < 0 {
		makerOrder.Side = "BUY"
		takerOrder.Side = "SELL"
	} else {
		makerOrder.Side = "SELL"
		takerOrder.Side = "BUY"
	}

	if err := s.doJSON(ctx, http.MethodPost, "/v1/orders", makerOrder, nil); err != nil {
		return err
	}
	if err := s.doJSON(ctx, http.MethodPost, "/v1/orders", takerOrder, nil); err != nil {
		return err
	}
	return nil
}

func (s *MatchingOrderSink) ensureFunding(ctx context.Context, symbol, makerUserID, takerUserID, baseAsset, quoteAsset string) error {
	s.mu.Lock()
	_, exists := s.funded[symbol]
	s.mu.Unlock()
	if exists {
		return nil
	}

	requests := []fundWalletRequest{
		{UserID: makerUserID, Asset: baseAsset, Amount: initialBotFunding},
		{UserID: makerUserID, Asset: quoteAsset, Amount: initialBotFunding},
		{UserID: takerUserID, Asset: baseAsset, Amount: initialBotFunding},
		{UserID: takerUserID, Asset: quoteAsset, Amount: initialBotFunding},
	}
	for _, req := range requests {
		if err := s.doJSON(ctx, http.MethodPost, "/v1/admin/wallets/fund", req, nil); err != nil {
			return err
		}
	}

	s.mu.Lock()
	s.funded[symbol] = struct{}{}
	s.mu.Unlock()
	return nil
}

func (s *MatchingOrderSink) nextOrderID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return fmt.Sprintf("sim-%d", s.seq)
}

func (s *MatchingOrderSink) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		message, _ := io.ReadAll(res.Body)
		trimmed := strings.TrimSpace(string(message))
		if trimmed == "" {
			trimmed = fmt.Sprintf("request failed: %s", res.Status)
		}
		return errors.New(trimmed)
	}

	if out == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}

type fundWalletRequest struct {
	UserID string `json:"userId"`
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
}

type orderPayload struct {
	ClientOrderID string `json:"clientOrderId"`
	UserID        string `json:"userId"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Type          string `json:"type"`
	Price         int64  `json:"price,omitempty"`
	Qty           int64  `json:"qty"`
}

func parseSymbol(symbol string) (string, string, error) {
	parts := strings.Split(symbol, "-")
	if len(parts) != 2 {
		return "", "", errors.New("symbol must be BASE-QUOTE format")
	}
	base := strings.TrimSpace(parts[0])
	quote := strings.TrimSpace(parts[1])
	if base == "" || quote == "" {
		return "", "", errors.New("symbol must be BASE-QUOTE format")
	}
	return strings.ToUpper(base), strings.ToUpper(quote), nil
}
