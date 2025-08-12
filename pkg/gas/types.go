package gas

import (
	"math/big"
)

// ChainClass indicates how the total fee for a transaction is calculated
type ChainClass string

const (
	ChainClassL1 = "L1"
	// TotalFees = gasLimit * (baseFeePerGas + priorityFeePerGas)
	// gasLimit <- eth_estimateGas
	// baseFeePerGas, priorityFeePerGas <- custom estimation
	ChainClassArbStack = "ArbStack"
	// TotalFees = gasLimit * (baseFeePerGas + priorityFeePerGas)
	// gasLimit <- eth_estimateGas
	// baseFeePerGas, priorityFeePerGas <- custom estimation
	ChainClassOPStack = "OPStack"
	// TotalFees = gasLimit * (baseFeePerGas + priorityFeePerGas) + l1Fee + operatorFee
	// gasLimit <- eth_estimateGas
	// baseFeePerGas, priorityFeePerGas <- custom estimation
	// l1Fee, operatorFee <- GasOracle
	ChainClassLineaStack = "LineaStack"
	// TotalFees = gasLimit * (baseFeePerGas + priorityFeePerGas)
	// gasLimit, baseFeePerGas, priorityFeePerGas <- linea_estimateGas
)

const (
	LowPriorityFeeIndex    = 0 // LowRewardPercentile
	MediumPriorityFeeIndex = 1 // MediumRewardPercentile
	HighPriorityFeeIndex   = 2 // HighRewardPercentile
)

type GasPrice struct {
	LowPriorityFeePerGas    *big.Int // Low priority fee per gas in wei
	MediumPriorityFeePerGas *big.Int // Medium priority fee per gas in wei
	HighPriorityFeePerGas   *big.Int // High priority fee per gas in wei
	BaseFeePerGas           *big.Int // Base fee per gas in wei
}

// FeeSuggestion represents a single fee suggestion level
type FeeSuggestion struct {
	MaxPriorityFeePerGas    *big.Int // Max priority fee per gas in wei
	MaxFeePerGas            *big.Int // Max fee per gas in wei
	MinBlocksUntilInclusion int      // Minimum number of blocks until inclusion
	MaxBlocksUntilInclusion int      // Maximum number of blocks until inclusion
	MinTimeUntilInclusion   float64  // Minimum time in seconds until inclusion
	MaxTimeUntilInclusion   float64  // Maximum time in seconds until inclusion
}

// FeeSuggestions represents the response from Infura's Gas API
type FeeSuggestions struct {
	Low                   FeeSuggestion // Low priority fee suggestion
	Medium                FeeSuggestion // Medium priority fee suggestion
	High                  FeeSuggestion // High priority fee suggestion
	EstimatedBaseFee      *big.Int      // Estimated base fee in wei
	PriorityFeeLowerBound *big.Int      // Recommended lower bound for priority fee per gas in wei
	PriorityFeeUpperBound *big.Int      // Recommended upper bound for priority fee per gas in wei
	NetworkCongestion     float64       // 0-1 scale. Only calculated for L1 chains
}

type TxSuggestions struct {
	FeeSuggestions *FeeSuggestions
	GasLimit       *big.Int
}

// SuggestionsConfig represents configuration for the gas suggestions
type SuggestionsConfig struct {
	ChainClass               ChainClass
	NetworkBlockTime         float64 // Average block time in seconds
	NetworkCongestionBlocks  int     // Number of blocks (from latest) to consider for network congestion estimation
	GasPriceEstimationBlocks int     // Number of blocks (from latest) to consider for gas price estimation
	LowRewardPercentile      float64 // Reward percentile for low priority fee
	MediumRewardPercentile   float64 // Reward percentile for medium priority fee
	HighRewardPercentile     float64 // Reward percentile for high priority fee
	BaseFeeMultiplier        float64 // Multiplier for base fee
}
