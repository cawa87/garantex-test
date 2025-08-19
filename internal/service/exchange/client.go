package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cawa87/garantex-test/internal/lib/logger/sl"
)

// Rate represents a currency exchange rate with ask/bid prices and timestamp
type Rate struct {
	Ask       float64   `json:"ask"`
	Bid       float64   `json:"bid"`
	Timestamp time.Time `json:"timestamp"`
}

// DepthResponse represents the response from Garantex depth API
type DepthResponse struct {
	Timestamp int64       `json:"timestamp"`
	Asks      []OrderBook `json:"asks"`
	Bids      []OrderBook `json:"bids"`
}

// OrderBook represents an order book entry with price and volume information
type OrderBook struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
	Factor string `json:"factor"`
	Type   string `json:"type"`
}

// Client represents the exchange API client for fetching rates
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *sl.Logger
}

// NewClient creates a new exchange client with the specified base URL and timeout
func NewClient(baseURL string, timeout time.Duration, logger *sl.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// GetRates fetches current USDT rates from Garantex exchange
func (c *Client) GetRates(ctx context.Context) (*Rate, error) {
	url := fmt.Sprintf("%s/api/v2/depth?market=btcusdt", c.baseURL)

	c.logger.Debug("Fetching rates from exchange", "url", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var depthResp DepthResponse
	if err := json.Unmarshal(body, &depthResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(depthResp.Bids) == 0 {
		return nil, fmt.Errorf("no bid prices available")
	}

	bid, err := parsePrice(depthResp.Bids[0].Price)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid price: %w", err)
	}

	ask := bid

	rate := &Rate{
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Now(),
	}

	c.logger.Info("Successfully fetched rates",
		"ask", rate.Ask,
		"bid", rate.Bid,
		"timestamp", rate.Timestamp)

	return rate, nil
}

// parsePrice converts string price to float64
func parsePrice(priceStr string) (float64, error) {
	var price float64
	_, err := fmt.Sscanf(priceStr, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("invalid price format: %s", priceStr)
	}
	return price, nil
}
