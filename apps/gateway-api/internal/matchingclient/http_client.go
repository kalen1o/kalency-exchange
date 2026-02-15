package matchingclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kalency/apps/gateway-api/internal/contracts"
)

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *HTTPClient) PlaceOrder(req contracts.PlaceOrderRequest) (contracts.OrderAck, error) {
	var ack contracts.OrderAck
	err := h.doJSON(http.MethodPost, "/v1/orders", req, &ack)
	return ack, err
}

func (h *HTTPClient) CancelOrder(userID, orderID string) (contracts.OrderAck, error) {
	var ack contracts.OrderAck
	path := fmt.Sprintf("/v1/orders/%s?userId=%s", url.PathEscape(orderID), url.QueryEscape(userID))
	err := h.doJSON(http.MethodDelete, path, nil, &ack)
	return ack, err
}

func (h *HTTPClient) OpenOrders(userID string) ([]contracts.Order, error) {
	var orders []contracts.Order
	err := h.doJSON(http.MethodGet, "/v1/orders/open/"+url.PathEscape(userID), nil, &orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (h *HTTPClient) Wallet(userID string) (contracts.Wallet, error) {
	var wallet contracts.Wallet
	err := h.doJSON(http.MethodGet, "/v1/wallet/"+url.PathEscape(userID), nil, &wallet)
	return wallet, err
}

func (h *HTTPClient) ListExecutions(symbol string, limit int) ([]contracts.Execution, error) {
	var executions []contracts.Execution
	path := fmt.Sprintf("/v1/markets/%s/trades?limit=%d", url.PathEscape(symbol), limit)
	err := h.doJSON(http.MethodGet, path, nil, &executions)
	if err != nil {
		return nil, err
	}
	return executions, nil
}

func (h *HTTPClient) ListOrderBook(symbol string, depth int) (contracts.OrderBookSnapshot, error) {
	var snapshot contracts.OrderBookSnapshot
	path := fmt.Sprintf("/v1/markets/%s/book?depth=%d", url.PathEscape(symbol), depth)
	err := h.doJSON(http.MethodGet, path, nil, &snapshot)
	return snapshot, err
}

func (h *HTTPClient) doJSON(method, path string, body any, out any) error {
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, h.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := h.client.Do(req)
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
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return err
	}
	return nil
}
