package infura

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetMainnetGasSuggestions(t *testing.T) {
	client := NewClient("test-api-key")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	suggestions, err := client.GetMainnetGasSuggestions(ctx)
	require.NoError(t, err)
	require.NotNil(t, suggestions)

	// Validate structure
	assert.NotEmpty(t, suggestions.Low.SuggestedMaxPriorityFeePerGas)
	assert.NotEmpty(t, suggestions.Low.SuggestedMaxFeePerGas)
	assert.NotEmpty(t, suggestions.Medium.SuggestedMaxPriorityFeePerGas)
	assert.NotEmpty(t, suggestions.Medium.SuggestedMaxFeePerGas)
	assert.NotEmpty(t, suggestions.High.SuggestedMaxPriorityFeePerGas)
	assert.NotEmpty(t, suggestions.High.SuggestedMaxFeePerGas)
	assert.NotEmpty(t, suggestions.EstimatedBaseFee)

	// Validate wait times are reasonable
	assert.True(t, suggestions.Low.MinWaitTimeEstimate >= 0)
	assert.True(t, suggestions.Low.MaxWaitTimeEstimate >= suggestions.Low.MinWaitTimeEstimate)
	assert.True(t, suggestions.Medium.MinWaitTimeEstimate >= 0)
	assert.True(t, suggestions.Medium.MaxWaitTimeEstimate >= suggestions.Medium.MinWaitTimeEstimate)
	assert.True(t, suggestions.High.MinWaitTimeEstimate >= 0)
	assert.True(t, suggestions.High.MaxWaitTimeEstimate >= suggestions.High.MinWaitTimeEstimate)

	// Validate network congestion is between 0 and 1
	assert.True(t, suggestions.NetworkCongestion >= 0 && suggestions.NetworkCongestion <= 1)

	t.Logf("Infura gas suggestions: %+v", suggestions)
}

func TestClient_GetGasSuggestions_DifferentNetworks(t *testing.T) {
	client := NewClient("test-api-key")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test Ethereum mainnet
	mainnetSuggestions, err := client.GetGasSuggestions(ctx, Ethereum)
	require.NoError(t, err)
	require.NotNil(t, mainnetSuggestions)

	// Test Polygon (if supported)
	if IsSupported(Polygon) {
		polygonSuggestions, err := client.GetGasSuggestions(ctx, Polygon)
		if err == nil {
			require.NotNil(t, polygonSuggestions)
			t.Logf("Polygon gas suggestions retrieved successfully")
		} else {
			t.Logf("Polygon gas suggestions failed (may not be available): %v", err)
		}
	}

	t.Logf("Mainnet suggestions retrieved successfully")
}

func TestClient_InvalidNetwork(t *testing.T) {
	client := NewClient("test-api-key")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with invalid network ID
	_, err := client.GetGasSuggestions(ctx, 99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "400") // Should return 400 for unsupported network
}

func TestNetworkHelpers(t *testing.T) {
	// Test GetNetworkName
	assert.Equal(t, "Ethereum", GetNetworkName(Ethereum))
	assert.Equal(t, "Polygon", GetNetworkName(Polygon))
	assert.Equal(t, "Network 99999", GetNetworkName(99999))

	// Test IsSupported
	assert.True(t, IsSupported(Ethereum))
	assert.True(t, IsSupported(Polygon))
	assert.False(t, IsSupported(99999))
}

func TestClient_WithCustomTimeout(t *testing.T) {
	client := NewClientWithTimeout("test-api-key", 5*time.Second)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	suggestions, err := client.GetMainnetGasSuggestions(ctx)
	require.NoError(t, err)
	require.NotNil(t, suggestions)

	assert.NotEmpty(t, suggestions.EstimatedBaseFee)
}

func BenchmarkClient_GetMainnetGasSuggestions(b *testing.B) {
	client := NewClient("test-api-key")
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetMainnetGasSuggestions(ctx)
		require.NoError(b, err)
	}
}
