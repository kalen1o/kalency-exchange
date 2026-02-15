package chartclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
		baseURL = "http://localhost:8085"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (h *HTTPClient) RenderChart(req contracts.ChartRenderRequest) (contracts.ChartRenderResponse, error) {
	var out contracts.ChartRenderResponse
	err := h.doJSON(http.MethodPost, "/v1/charts/render", req, &out)
	return out, err
}

func (h *HTTPClient) doJSON(method, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, h.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

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
	return json.NewDecoder(res.Body).Decode(out)
}
