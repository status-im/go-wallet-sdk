package gas_test

import (
	"math/big"
	"testing"

	"github.com/status-im/go-wallet-sdk/pkg/gas"
	"github.com/stretchr/testify/assert"
)

func TestWeiRepresentation(t *testing.T) {
	// Test that our structs properly represent values in wei

	// Create a FeeSuggestion with known wei values
	suggestion := gas.FeeSuggestion{
		SuggestedMaxPriorityFeePerGas: big.NewInt(2000000000),  // 2 gwei in wei
		SuggestedMaxFeePerGas:         big.NewInt(22000000000), // 22 gwei in wei
		MinWaitTimeEstimate:           15,
		MaxWaitTimeEstimate:           60,
	}

	// Verify the values are stored as wei
	assert.Equal(t, int64(2000000000), suggestion.SuggestedMaxPriorityFeePerGas.Int64())
	assert.Equal(t, int64(22000000000), suggestion.SuggestedMaxFeePerGas.Int64())

	// Test conversion to gwei for display
	priorityFeeGwei := float64(suggestion.SuggestedMaxPriorityFeePerGas.Int64()) / 1e9
	maxFeeGwei := float64(suggestion.SuggestedMaxFeePerGas.Int64()) / 1e9

	assert.Equal(t, 2.0, priorityFeeGwei)
	assert.Equal(t, 22.0, maxFeeGwei)

	// Create a FeeSuggestions with known wei values
	suggestions := gas.FeeSuggestions{
		Low: gas.FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: big.NewInt(1000000000),  // 1 gwei
			SuggestedMaxFeePerGas:         big.NewInt(21000000000), // 21 gwei
		},
		Medium: gas.FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: big.NewInt(2000000000),  // 2 gwei
			SuggestedMaxFeePerGas:         big.NewInt(22000000000), // 22 gwei
		},
		High: gas.FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: big.NewInt(5000000000),  // 5 gwei
			SuggestedMaxFeePerGas:         big.NewInt(25000000000), // 25 gwei
		},
		EstimatedBaseFee:  big.NewInt(20000000000), // 20 gwei in wei
		NetworkCongestion: 0.75,
	}

	// Verify all values are stored as wei
	assert.Equal(t, int64(1000000000), suggestions.Low.SuggestedMaxPriorityFeePerGas.Int64())
	assert.Equal(t, int64(21000000000), suggestions.Low.SuggestedMaxFeePerGas.Int64())
	assert.Equal(t, int64(2000000000), suggestions.Medium.SuggestedMaxPriorityFeePerGas.Int64())
	assert.Equal(t, int64(22000000000), suggestions.Medium.SuggestedMaxFeePerGas.Int64())
	assert.Equal(t, int64(5000000000), suggestions.High.SuggestedMaxPriorityFeePerGas.Int64())
	assert.Equal(t, int64(25000000000), suggestions.High.SuggestedMaxFeePerGas.Int64())
	assert.Equal(t, int64(20000000000), suggestions.EstimatedBaseFee.Int64())

	// Test conversion to gwei for all values
	lowPriorityGwei := float64(suggestions.Low.SuggestedMaxPriorityFeePerGas.Int64()) / 1e9
	mediumMaxGwei := float64(suggestions.Medium.SuggestedMaxFeePerGas.Int64()) / 1e9
	baseFeeGwei := float64(suggestions.EstimatedBaseFee.Int64()) / 1e9

	assert.Equal(t, 1.0, lowPriorityGwei)
	assert.Equal(t, 22.0, mediumMaxGwei)
	assert.Equal(t, 20.0, baseFeeGwei)
}

func TestWeiMathOperations(t *testing.T) {
	// Test that we can perform mathematical operations directly on wei values

	baseFee := big.NewInt(20000000000)    // 20 gwei
	priorityFee := big.NewInt(2000000000) // 2 gwei

	// Calculate max fee per gas (base fee + priority fee)
	maxFee := new(big.Int).Add(baseFee, priorityFee)

	// Verify the result
	assert.Equal(t, int64(22000000000), maxFee.Int64()) // 22 gwei in wei

	// Test multiplication for buffer calculations
	bufferedBaseFee := new(big.Int).Mul(baseFee, big.NewInt(1125)) // 112.5% multiplier
	bufferedBaseFee.Div(bufferedBaseFee, big.NewInt(1000))

	// Verify the buffered base fee (20 * 1.125 = 22.5 gwei)
	assert.Equal(t, int64(22500000000), bufferedBaseFee.Int64())

	// Test comparison operations
	assert.True(t, priorityFee.Cmp(big.NewInt(1000000000)) > 0) // 2 gwei > 1 gwei
	assert.True(t, maxFee.Cmp(baseFee) > 0)                     // max fee > base fee
}

func TestWeiPrecision(t *testing.T) {
	// Test that wei representation maintains precision for small values

	// Very small priority fee (0.1 gwei = 100,000,000 wei)
	smallFee := big.NewInt(100000000)

	// Convert to gwei and back
	gweiFloat := float64(smallFee.Int64()) / 1e9
	backToWei := int64(gweiFloat * 1e9)

	assert.Equal(t, 0.1, gweiFloat)
	assert.Equal(t, smallFee.Int64(), backToWei)

	// Test with fractional gwei values (1.5 gwei = 1,500,000,000 wei)
	fractionalFee := big.NewInt(1500000000)
	fractionalGwei := float64(fractionalFee.Int64()) / 1e9

	assert.Equal(t, 1.5, fractionalGwei)
}
