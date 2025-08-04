package gas_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/status-im/go-wallet-sdk/pkg/gas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeeSuggestionJSONSerialization(t *testing.T) {
	// Create a FeeSuggestion with big.Int values
	suggestion := gas.FeeSuggestion{
		SuggestedMaxPriorityFeePerGas: big.NewInt(2000000000),  // 2 gwei in wei
		SuggestedMaxFeePerGas:         big.NewInt(22000000000), // 22 gwei in wei
		MinWaitTimeEstimate:           15,
		MaxWaitTimeEstimate:           60,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(suggestion)
	require.NoError(t, err)

	t.Logf("JSON output: %s", string(jsonData))

	// Test JSON unmarshaling
	var unmarshaled gas.FeeSuggestion
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify the values are preserved
	assert.Equal(t, suggestion.SuggestedMaxPriorityFeePerGas.String(), unmarshaled.SuggestedMaxPriorityFeePerGas.String())
	assert.Equal(t, suggestion.SuggestedMaxFeePerGas.String(), unmarshaled.SuggestedMaxFeePerGas.String())
	assert.Equal(t, suggestion.MinWaitTimeEstimate, unmarshaled.MinWaitTimeEstimate)
	assert.Equal(t, suggestion.MaxWaitTimeEstimate, unmarshaled.MaxWaitTimeEstimate)
}

func TestFeeSuggestionsJSONSerialization(t *testing.T) {
	// Create a FeeSuggestions with big.Int values
	suggestions := gas.FeeSuggestions{
		Low: gas.FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: big.NewInt(1000000000),  // 1 gwei
			SuggestedMaxFeePerGas:         big.NewInt(21000000000), // 21 gwei
			MinWaitTimeEstimate:           60,
			MaxWaitTimeEstimate:           300,
		},
		Medium: gas.FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: big.NewInt(2000000000),  // 2 gwei
			SuggestedMaxFeePerGas:         big.NewInt(22000000000), // 22 gwei
			MinWaitTimeEstimate:           15,
			MaxWaitTimeEstimate:           60,
		},
		High: gas.FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: big.NewInt(5000000000),  // 5 gwei
			SuggestedMaxFeePerGas:         big.NewInt(25000000000), // 25 gwei
			MinWaitTimeEstimate:           0,
			MaxWaitTimeEstimate:           15,
		},
		EstimatedBaseFee:  big.NewInt(20000000000), // 20 gwei in wei
		NetworkCongestion: 0.75,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(suggestions)
	require.NoError(t, err)

	t.Logf("JSON output: %s", string(jsonData))

	// Test JSON unmarshaling
	var unmarshaled gas.FeeSuggestions
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify the values are preserved
	assert.Equal(t, suggestions.Low.SuggestedMaxPriorityFeePerGas.String(), unmarshaled.Low.SuggestedMaxPriorityFeePerGas.String())
	assert.Equal(t, suggestions.Medium.SuggestedMaxFeePerGas.String(), unmarshaled.Medium.SuggestedMaxFeePerGas.String())
	assert.Equal(t, suggestions.High.MinWaitTimeEstimate, unmarshaled.High.MinWaitTimeEstimate)
	assert.Equal(t, suggestions.EstimatedBaseFee, unmarshaled.EstimatedBaseFee)
	assert.Equal(t, suggestions.NetworkCongestion, unmarshaled.NetworkCongestion)
}
