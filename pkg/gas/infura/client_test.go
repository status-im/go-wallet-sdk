package infura_test

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

//go:embed test/suggested_gas_fees.json
var suggestedGasFeesJSON string

func TestClient_GetGasSuggestions_Success(t *testing.T) {
	// Create a test server that returns the embedded JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/networks/1/suggestedGasFees", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(suggestedGasFeesJSON))
		require.NoError(t, err)
	}))
	defer server.Close()

	// Create client with test server URL
	client := infura.NewClientWithBaseURL(http.DefaultClient, server.URL)

	// Call GetGasSuggestions
	ctx := context.Background()
	response, err := client.GetGasSuggestions(ctx, 1)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify low tier
	assert.Equal(t, "0.05", response.Low.SuggestedMaxPriorityFeePerGas)
	assert.Equal(t, "16.334026964", response.Low.SuggestedMaxFeePerGas)
	assert.Equal(t, 15000, response.Low.MinWaitTimeEstimate)
	assert.Equal(t, 30000, response.Low.MaxWaitTimeEstimate)

	// Verify medium tier
	assert.Equal(t, "0.1", response.Medium.SuggestedMaxPriorityFeePerGas)
	assert.Equal(t, "22.083436402", response.Medium.SuggestedMaxFeePerGas)
	assert.Equal(t, 15000, response.Medium.MinWaitTimeEstimate)
	assert.Equal(t, 45000, response.Medium.MaxWaitTimeEstimate)

	// Verify high tier
	assert.Equal(t, "0.3", response.High.SuggestedMaxPriorityFeePerGas)
	assert.Equal(t, "27.982845839", response.High.SuggestedMaxFeePerGas)
	assert.Equal(t, 15000, response.High.MinWaitTimeEstimate)
	assert.Equal(t, 60000, response.High.MaxWaitTimeEstimate)

	// Verify other fields
	assert.Equal(t, "16.284026964", response.EstimatedBaseFee)
	assert.Equal(t, 0.5125, response.NetworkCongestion)
	assert.Equal(t, []string{"0", "3"}, response.LatestPriorityFeeRange)
	assert.Equal(t, []string{"0.000000001", "89"}, response.HistoricalPriorityFeeRange)
	assert.Equal(t, []string{"13.773088584", "29.912845463"}, response.HistoricalBaseFeeRange)
	assert.Equal(t, "down", response.PriorityFeeTrend)
	assert.Equal(t, "up", response.BaseFeeTrend)
}

func TestClient_GetGasSuggestions_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Internal Server Error"))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := infura.NewClientWithBaseURL(http.DefaultClient, server.URL)

	ctx := context.Background()
	response, err := client.GetGasSuggestions(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "API returned status 500")
}

func TestClient_GetGasSuggestions_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("invalid json"))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := infura.NewClientWithBaseURL(http.DefaultClient, server.URL)

	ctx := context.Background()
	response, err := client.GetGasSuggestions(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to decode response")
}
