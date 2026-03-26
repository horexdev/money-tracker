package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RateProvider fetches exchange rates for a given base currency.
type RateProvider interface {
	FetchRates(ctx context.Context, base string) (map[string]float64, error)
}

// RateAPIProvider calls the open.er-api.com free API.
type RateAPIProvider struct {
	client *http.Client
}

// NewRateAPIProvider creates a provider with a 10-second timeout.
func NewRateAPIProvider() *RateAPIProvider {
	return &RateAPIProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type erAPIResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

// FetchRates returns exchange rates relative to the given base currency.
func (p *RateAPIProvider) FetchRates(ctx context.Context, base string) (map[string]float64, error) {
	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", base)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch rates for %s: %w", base, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from exchange API", resp.StatusCode)
	}

	var body erAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if body.Result != "success" {
		return nil, fmt.Errorf("exchange API returned result=%q", body.Result)
	}

	return body.Rates, nil
}
