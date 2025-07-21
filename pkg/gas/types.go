package gas

import (
	"math/big"
)

// FeeSuggestion represents a single fee suggestion level
type FeeSuggestion struct {
	SuggestedMaxPriorityFeePerGas *big.Int `json:"suggestedMaxPriorityFeePerGas"` // Fee per gas in wei
	SuggestedMaxFeePerGas         *big.Int `json:"suggestedMaxFeePerGas"`         // Max fee per gas in wei
	MinWaitTimeEstimate           int      `json:"minWaitTimeEstimate"`           // Minimum wait time in seconds
	MaxWaitTimeEstimate           int      `json:"maxWaitTimeEstimate"`           // Maximum wait time in seconds
}

// FeeSuggestions represents the response from Infura's Gas API
type FeeSuggestions struct {
	Low               FeeSuggestion `json:"low"`
	Medium            FeeSuggestion `json:"medium"`
	High              FeeSuggestion `json:"high"`
	EstimatedBaseFee  *big.Int      `json:"estimatedBaseFee"` // Estimated base fee in wei
	NetworkCongestion float64       `json:"networkCongestion"`
}

// EstimatorConfig represents configuration for the gas estimator
type EstimatorConfig struct {
	LegacyBlocks      int     `json:"legacyBlocks"`
	PercentileBlocks  int     `json:"percentileBlocks"`
	LowPercentile     float64 `json:"lowPercentile"`
	MediumPercentile  float64 `json:"mediumPercentile"`
	HighPercentile    float64 `json:"highPercentile"`
	BaseFeeMultiplier float64 `json:"baseFeeMultiplier"`
	NetworkBlockTime  float64 `json:"networkBlockTime"` // Average block time in seconds
}

// EstimatedInclusionResult contains the estimated inclusion time and block count
type EstimatedInclusionResult struct {
	MinBlocks      int    `json:"minBlocks"`      // Minimum number of blocks until inclusion
	MaxBlocks      int    `json:"maxBlocks"`      // Maximum number of blocks until inclusion
	MinTimeSeconds int    `json:"minTimeSeconds"` // Minimum time in seconds until inclusion
	MaxTimeSeconds int    `json:"maxTimeSeconds"` // Maximum time in seconds until inclusion
	Confidence     string `json:"confidence"`     // Confidence level (high, medium, low)
}
