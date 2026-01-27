package gas

import (
	"math/big"
)

// FeeModel specifies which fee market rules the chain follows.
//
// - FeeModelEIP1559: dynamic fee transactions (maxFeePerGas + maxPriorityFeePerGas)
// - FeeModelLegacy: legacy transactions (gasPrice)
type FeeModel string

const (
	FeeModelEIP1559 FeeModel = "EIP1559"
	FeeModelLegacy  FeeModel = "Legacy"
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

// Fee represents a single fee suggestion level
type Fee struct {
	// GasPrice is used for legacy (non-EIP-1559) chains (when GasPrice is set, MaxFeePerGas/MaxPriorityFeePerGas are expected to be nil).
	GasPrice *big.Int // Gas price in wei

	MaxPriorityFeePerGas *big.Int // Max priority fee per gas in wei
	MaxFeePerGas         *big.Int // Max fee per gas in wei
}

type Inclusion struct {
	MinBlocksUntilInclusion int     // Minimum number of blocks until inclusion
	MaxBlocksUntilInclusion int     // Maximum number of blocks until inclusion
	MinTimeUntilInclusion   float64 // Minimum time in seconds until inclusion
	MaxTimeUntilInclusion   float64 // Maximum time in seconds until inclusion
}

// FeeSuggestions represents the response from Infura's Gas API
type FeeSuggestions struct {
	FeeModel FeeModel // Fee model used for the suggestions (EIP1559 or Legacy)

	Low             Fee       // Low priority fee suggestion
	LowInclusion    Inclusion // Low priority fee inclusion
	Medium          Fee       // Medium priority fee suggestion
	MediumInclusion Inclusion // Medium priority fee inclusion
	High            Fee       // High priority fee suggestion
	HighInclusion   Inclusion // High priority fee inclusion

	// Set if FeeModel is FeeModelLegacy
	GasPrice *big.Int // Gas price in wei

	// Set if FeeModel is FeeModelEIP1559
	EstimatedBaseFee      *big.Int // Estimated base fee in wei
	PriorityFeeLowerBound *big.Int // Recommended lower bound for priority fee per gas in wei
	PriorityFeeUpperBound *big.Int // Recommended upper bound for priority fee per gas in wei

	NetworkCongestion float64 // 0-1 scale. Only calculated for L1 chains
}

type TxSuggestions struct {
	FeeSuggestions *FeeSuggestions
	GasLimit       *big.Int
}

// SuggestionsConfig represents configuration for the gas suggestions
type SuggestionsConfig struct {
	NetworkCongestionBlocks  int // Number of blocks (from latest) to consider for network congestion estimation
	GasPriceEstimationBlocks int // Number of blocks (from latest) to consider for gas price estimation

	LowRewardPercentile     float64 // Reward percentile for low priority fee
	MediumRewardPercentile  float64 // Reward percentile for medium priority fee
	HighRewardPercentile    float64 // Reward percentile for high priority fee
	LowBaseFeeMultiplier    float64 // Multiplier for base fee for low level
	MediumBaseFeeMultiplier float64 // Multiplier for base fee for medium level
	HighBaseFeeMultiplier   float64 // Multiplier for base fee for high level

	LowBaseFeeCongestionMultiplier    float64 // (Only ChainClassL1) A factor of (1 + congestion * LowBaseFeeCongestionMultiplier) will be applied to the base fee for the Low level
	MediumBaseFeeCongestionMultiplier float64 // (Only ChainClassL1) A factor of (1 + congestion * MediumBaseFeeCongestionMultiplier) will be applied to the base fee for the Medium level
	HighBaseFeeCongestionMultiplier   float64 // (Only ChainClassL1) A factor of (1 + congestion * HighBaseFeeCongestionMultiplier) will be applied to the base fee for the High level

	// Legacy gasPrice multipliers (used when FeeModelLegacy if no Transactions GasPrice are available).
	LowGasPriceMultiplier    float64 // Multiplier for gasPrice for the Low level
	MediumGasPriceMultiplier float64 // Multiplier for gasPrice for the Medium level
	HighGasPriceMultiplier   float64 // Multiplier for gasPrice for the High level
}

// ChainParameters includes chain-specific parameters
type ChainParameters struct {
	ChainClass       ChainClass
	NetworkBlockTime float64 // Average block time in seconds
	FeeModel         FeeModel
}
