package gas

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

// EstimateTransactionInclusion estimates when a transaction will be included based on its fee
// and current network conditions
func (e *Estimator) EstimateTransactionInclusion(
	ctx context.Context,
	priorityFee *big.Int,
	maxFeePerGas *big.Int,
) (*EstimatedInclusionResult, error) {
	// Get latest block to check network conditions
	latestBlock, err := e.getLatestBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	// Get fee history to analyze recent priority fee trends
	feeHistory, err := e.getFeeHistory(ctx, e.config.PercentileBlocks)
	if err != nil {
		// If fee history is not available, use a simplified estimation
		return e.estimateInclusionSimplified(priorityFee, latestBlock)
	}

	// Calculate network congestion
	networkCongestion := e.calculateNetworkCongestionFromHistory(feeHistory)

	// Determine average block time for the network (default to 12s for Ethereum)
	blockTime := 12.0 // seconds
	if e.config.NetworkBlockTime > 0 {
		blockTime = e.config.NetworkBlockTime
	}

	// Convert priority fee to gwei for easier comparison
	priorityFeeGwei := new(big.Float).Quo(
		new(big.Float).SetInt(priorityFee),
		new(big.Float).SetInt(big.NewInt(1e9)),
	)
	priorityFeeFloat, _ := priorityFeeGwei.Float64()

	// Analyze recent priority fees to determine competitiveness
	competitiveness := e.calculateFeeCompetitiveness(priorityFeeFloat, feeHistory)

	// Base wait time in blocks
	var baseWaitBlocks float64
	switch competitiveness {
	case "high":
		baseWaitBlocks = 1.0
	case "medium_high":
		baseWaitBlocks = 1.5
	case "medium":
		baseWaitBlocks = 2.0
	case "low":
		baseWaitBlocks = 4.0
	case "very_low":
		baseWaitBlocks = 8.0
	default:
		baseWaitBlocks = 2.0
	}

	// Adjust wait time based on network congestion
	congestionFactor := 1.0
	if networkCongestion > 0.7 {
		// Exponential increase in wait time as congestion approaches 100%
		congestionFactor = math.Pow(1.0+networkCongestion, 2)
	}

	// Calculate estimated wait time in blocks
	waitBlocks := baseWaitBlocks * congestionFactor

	// Add variability for min/max estimates
	minBlocks := int(math.Max(1, math.Floor(waitBlocks*0.7)))
	maxBlocks := int(math.Ceil(waitBlocks * 1.5))

	// Convert to time estimates
	minTimeSeconds := int(float64(minBlocks) * blockTime)
	maxTimeSeconds := int(float64(maxBlocks) * blockTime)

	// Determine confidence level
	confidence := "medium"
	if competitiveness == "high" && networkCongestion < 0.8 {
		confidence = "high"
	} else if competitiveness == "low" || networkCongestion > 0.9 {
		confidence = "low"
	}

	return &EstimatedInclusionResult{
		MinBlocks:      minBlocks,
		MaxBlocks:      maxBlocks,
		MinTimeSeconds: minTimeSeconds,
		MaxTimeSeconds: maxTimeSeconds,
		Confidence:     confidence,
	}, nil
}

// calculateFeeCompetitiveness determines how competitive a priority fee is
// compared to recent transactions
func (e *Estimator) calculateFeeCompetitiveness(priorityFeeGwei float64, feeHistory *ethclient.FeeHistory) string {
	// Extract all priority fees from recent blocks
	var recentPriorityFees []float64
	for _, blockRewards := range feeHistory.Reward {
		for _, reward := range blockRewards {
			if reward != nil && reward.Sign() > 0 {
				recentPriorityFees = append(recentPriorityFees, float64(reward.Uint64())/1e9)
			}
		}
	}

	if len(recentPriorityFees) == 0 {
		// No data available, assume medium competitiveness
		return "medium"
	}

	// Sort fees to calculate percentiles
	sort.Float64s(recentPriorityFees)

	// Calculate percentiles
	p25 := percentileFloat64(recentPriorityFees, 25)
	p50 := percentileFloat64(recentPriorityFees, 50)
	p75 := percentileFloat64(recentPriorityFees, 75)
	p90 := percentileFloat64(recentPriorityFees, 90)

	// Determine competitiveness based on where the fee falls
	if priorityFeeGwei >= p90 {
		return "high"
	} else if priorityFeeGwei >= p75 {
		return "medium_high"
	} else if priorityFeeGwei >= p50 {
		return "medium"
	} else if priorityFeeGwei >= p25 {
		return "low"
	} else {
		return "very_low"
	}
}

// percentileFloat64 calculates the percentile value from sorted data
func percentileFloat64(sortedData []float64, percentile int) float64 {
	if len(sortedData) == 0 {
		return 0
	}
	index := (len(sortedData) * percentile) / 100
	if index >= len(sortedData) {
		index = len(sortedData) - 1
	}
	return sortedData[index]
}

// estimateInclusionSimplified provides a simplified estimation when fee history is not available
func (e *Estimator) estimateInclusionSimplified(priorityFee *big.Int, latestBlock *ethclient.BlockWithFullTxs) (*EstimatedInclusionResult, error) {
	// Convert priority fee to gwei
	priorityFeeGwei := new(big.Float).Quo(
		new(big.Float).SetInt(priorityFee),
		new(big.Float).SetInt(big.NewInt(1e9)),
	)
	priorityFeeFloat, _ := priorityFeeGwei.Float64()

	// Default block time (12s for Ethereum)
	blockTime := 12.0
	if e.config.NetworkBlockTime > 0 {
		blockTime = e.config.NetworkBlockTime
	}

	// Simple heuristic based on priority fee
	var minBlocks, maxBlocks int
	var confidence string

	if priorityFeeFloat >= 3.0 {
		minBlocks = 1
		maxBlocks = 2
		confidence = "high"
	} else if priorityFeeFloat >= 1.5 {
		minBlocks = 1
		maxBlocks = 3
		confidence = "medium"
	} else if priorityFeeFloat >= 0.5 {
		minBlocks = 2
		maxBlocks = 5
		confidence = "medium"
	} else {
		minBlocks = 3
		maxBlocks = 10
		confidence = "low"
	}

	// Calculate time estimates
	minTimeSeconds := int(float64(minBlocks) * blockTime)
	maxTimeSeconds := int(float64(maxBlocks) * blockTime)

	return &EstimatedInclusionResult{
		MinBlocks:      minBlocks,
		MaxBlocks:      maxBlocks,
		MinTimeSeconds: minTimeSeconds,
		MaxTimeSeconds: maxTimeSeconds,
		Confidence:     confidence,
	}, nil
}

// getTimeEstimatesForFee calculates time estimates for a given priority fee using inclusion estimation
func (e *Estimator) getTimeEstimatesForFee(ctx context.Context, priorityFee *big.Int, maxFeePerGas *big.Int) (int, int) {
	// Try to get accurate estimates using our inclusion estimation
	result, err := e.EstimateTransactionInclusion(ctx, priorityFee, maxFeePerGas)
	if err == nil {
		return result.MinTimeSeconds, result.MaxTimeSeconds
	}

	// Fallback to simple heuristics if estimation fails or no client available
	priorityFeeGwei := new(big.Float).Quo(
		new(big.Float).SetInt(priorityFee),
		new(big.Float).SetInt(big.NewInt(1e9)),
	)
	priorityFeeFloat, _ := priorityFeeGwei.Float64()

	if priorityFeeFloat >= 3.0 {
		return 0, 15 // High fee: immediate to 15 seconds
	} else if priorityFeeFloat >= 1.5 {
		return 15, 60 // Medium fee: 15 seconds to 1 minute
	} else {
		return 60, 300 // Low fee: 1 to 5 minutes
	}
}
