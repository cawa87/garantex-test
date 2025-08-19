package exchange

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cawa87/garantex-test/internal/lib/logger/sl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	logger, err := sl.New("info")
	require.NoError(t, err)

	client := NewClient("https://test.com", 10*time.Second, logger)

	assert.NotNil(t, client)
	assert.Equal(t, "https://test.com", client.baseURL)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestGetRates_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/depth", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Return mock response
		response := `{
			"timestamp": 1755631475,
			"asks": [],
			"bids": [
				{
					"price": "100.40",
					"volume": "1.5",
					"amount": "150.60",
					"factor": "0.226",
					"type": "limit"
				}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	logger, err := sl.New("info")
	require.NoError(t, err)

	client := NewClient(server.URL, 10*time.Second, logger)

	ctx := context.Background()
	rate, err := client.GetRates(ctx)

	require.NoError(t, err)
	assert.NotNil(t, rate)
	assert.Equal(t, 100.40, rate.Ask)
	assert.Equal(t, 100.40, rate.Bid)
	assert.WithinDuration(t, time.Now(), rate.Timestamp, 2*time.Second)
}

func TestGetRates_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{"timestamp": 1755631475, "asks": [], "bids": []}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	logger, err := sl.New("info")
	require.NoError(t, err)

	client := NewClient(server.URL, 10*time.Second, logger)

	ctx := context.Background()
	_, err = client.GetRates(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no bid prices available")
}

func TestGetRates_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	logger, err := sl.New("info")
	require.NoError(t, err)

	client := NewClient(server.URL, 10*time.Second, logger)

	ctx := context.Background()
	_, err = client.GetRates(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal response")
}

func TestGetRates_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	logger, err := sl.New("info")
	require.NoError(t, err)

	client := NewClient(server.URL, 10*time.Second, logger)

	ctx := context.Background()
	_, err = client.GetRates(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}

func TestGetRates_InvalidPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"timestamp": 1755631475,
			"asks": [],
			"bids": [
				{
					"price": "invalid",
					"volume": "1.5",
					"amount": "150.60",
					"factor": "0.226",
					"type": "limit"
				}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	logger, err := sl.New("info")
	require.NoError(t, err)

	client := NewClient(server.URL, 10*time.Second, logger)

	ctx := context.Background()
	_, err = client.GetRates(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse bid price")
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name     string
		priceStr string
		expected float64
		hasError bool
	}{
		{"valid price", "100.50", 100.50, false},
		{"integer price", "100", 100.0, false},
		{"zero price", "0", 0.0, false},
		{"invalid price", "invalid", 0.0, true},
		{"empty string", "", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := parsePrice(tt.priceStr)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, price)
			}
		})
	}
}
