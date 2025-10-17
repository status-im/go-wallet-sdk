package gas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateNetworkCongestion_EdgeCases(t *testing.T) {
	// Test with very high gas usage (near 100%)
	highGasUsage := []float64{0.95, 0.98, 0.99, 0.97, 0.96}
	result := calculateNetworkCongestion(highGasUsage)
	assert.GreaterOrEqual(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
	assert.Greater(t, result, 0.5) // Should indicate high congestion

	// Test with very low gas usage (near 0%)
	lowGasUsage := []float64{0.05, 0.02, 0.01, 0.03, 0.04}
	result = calculateNetworkCongestion(lowGasUsage)
	assert.GreaterOrEqual(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
	assert.Less(t, result, 0.5) // Should indicate low congestion

	// Test with exactly 50% gas usage
	mediumGasUsage := []float64{0.5, 0.5, 0.5, 0.5, 0.5}
	result = calculateNetworkCongestion(mediumGasUsage)

	assert.GreaterOrEqual(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)

	// Test with alternating high/low gas usage
	alternatingGasUsage := []float64{0.1, 0.9, 0.2, 0.8, 0.3, 0.7}
	result = calculateNetworkCongestion(alternatingGasUsage)
	assert.GreaterOrEqual(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
}
