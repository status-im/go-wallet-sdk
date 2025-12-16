package infura

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client represents an Infura Gas API client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Infura Gas API client
func NewClient(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    "https://gas-api.metaswap.codefi.network",
	}
}

// NewClientWithBaseURL creates a new Infura Gas API client with a custom base URL.
// This is primarily useful for testing.
func NewClientWithBaseURL(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// GetGasSuggestions retrieves gas fee suggestions from Infura's Gas API
func (c *Client) GetGasSuggestions(ctx context.Context, networkID int) (*GasResponse, error) {
	url := fmt.Sprintf("%s/networks/%d/suggestedGasFees", c.baseURL, networkID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var gasResponse GasResponse
	if err := json.NewDecoder(resp.Body).Decode(&gasResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &gasResponse, nil
}

// Close closes the HTTP client (currently a no-op, but provided for consistency)
func (c *Client) Close() error {
	// HTTP client doesn't need explicit closing, but we provide this for consistency
	return nil
}
