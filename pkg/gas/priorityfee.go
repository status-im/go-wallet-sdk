package gas

import (
	"math/big"
	"slices"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

// calculatePriorityFeeFromHistory calculates priority fee for a given percentile from fee history
func calculatePriorityFeeFromHistory(feeHistory *ethclient.FeeHistory, nBlocks int, rewardsPercentileIndex int) *big.Int {
	priorityFees := extractPriorityFeesFromHistory(feeHistory, rewardsPercentileIndex)

	startIdx := max(len(priorityFees)-nBlocks, 0)
	return calculatePriorityFee(priorityFees[startIdx:])
}

// extractPriorityFeesFromHistory extracts priority fees for a given rewardsPercentileIndex from fee history
func extractPriorityFeesFromHistory(feeHistory *ethclient.FeeHistory, rewardsPercentileIndex int) []*big.Int {
	priorityFees := make([]*big.Int, 0)
	for _, blockRewards := range feeHistory.Reward {
		if rewardsPercentileIndex < len(blockRewards) && blockRewards[rewardsPercentileIndex] != nil {
			priorityFees = append(priorityFees, new(big.Int).Set(blockRewards[rewardsPercentileIndex]))
		}
	}

	return priorityFees
}

func calculatePriorityFee(priorityFees []*big.Int) *big.Int {
	if len(priorityFees) == 0 {
		return big.NewInt(1000000000) // 1 gwei fallback
	}

	// Calculate median of the collected priority fees for stability
	sortedFees := make([]*big.Int, len(priorityFees))
	copy(sortedFees, priorityFees)
	slices.SortFunc(sortedFees, func(a, b *big.Int) int {
		return a.Cmp(b)
	})

	medianIndex := len(sortedFees) / 2
	medianFee := sortedFees[medianIndex]

	return medianFee
}
