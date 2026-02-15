package marketsimclient

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
)

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8082"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *HTTPClient) StartSimulator() (map[string]any, error) {
	var out map[string]any
	err := h.doJSON(http.MethodPost, "/v1/admin/sim/start", nil, &out)
	return out, err
}

func (h *HTTPClient) StopSimulator() (map[string]any, error) {
	var out map[string]any
	err := h.doJSON(http.MethodPost, "/v1/admin/sim/stop", nil, &out)
	return out, err
}

func (h *HTTPClient) SetVolatility(volatility float64) (map[string]any, error) {
	var out map[string]any
	body := map[string]float64{"volatility": volatility}
	err := h.doJSON(http.MethodPost, "/v1/admin/sim/volatility-profile", body, &out)
	return out, err
}

func (h *HTTPClient) PauseSymbol(symbol string) (map[string]any, error) {
	var out map[string]any
	path := fmt.Sprintf("/v1/admin/symbols/%s/pause", url.PathEscape(strings.TrimSpace(symbol)))
	err := h.doJSON(http.MethodPost, path, nil, &out)
	return out, err
}

func (h *HTTPClient) ResumeSymbol(symbol string) (map[string]any, error) {
	var out map[string]any
	path := fmt.Sprintf("/v1/admin/symbols/%s/resume", url.PathEscape(strings.TrimSpace(symbol)))
	err := h.doJSON(http.MethodPost, path, nil, &out)
	return out, err
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
