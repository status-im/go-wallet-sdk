package gas

import (
	"math/big"
	"slices"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

// For any chain except Linea Stack
func fillInclusions(txSuggestions *TxSuggestions, feeHistory *ethclient.FeeHistory, avgBlockTime float64) {
	sortedBaseFees := make([]*big.Int, len(feeHistory.BaseFeePerGas[:len(feeHistory.BaseFeePerGas)-2]))
	copy(sortedBaseFees, feeHistory.BaseFeePerGas[:len(feeHistory.BaseFeePerGas)-2])
	slices.SortFunc(sortedBaseFees, func(a, b *big.Int) int {
		return a.Cmp(b)
	})

	sortedMediumPriorityFees := extractPriorityFeesFromHistory(feeHistory, MediumPriorityFeeIndex)
	slices.SortFunc(sortedMediumPriorityFees, func(a, b *big.Int) int {
		return a.Cmp(b)
	})

	feeSuggestions := []*FeeSuggestion{
		&txSuggestions.FeeSuggestions.Low,
		&txSuggestions.FeeSuggestions.Medium,
		&txSuggestions.FeeSuggestions.High,
	}
	for _, feeSuggestion := range feeSuggestions {
		fillInclusion(feeSuggestion, sortedBaseFees, sortedMediumPriorityFees, avgBlockTime)
	}
}

func fillInclusion(feeSugestion *FeeSuggestion, sortedBaseFees []*big.Int, sortedMediumPriorityFees []*big.Int, avgBlockTime float64) {
	minBlocks, maxBlocks := estimateBlocksUntilInclusion(feeSugestion.MaxPriorityFeePerGas, feeSugestion.MaxFeePerGas, sortedBaseFees, sortedMediumPriorityFees)

	feeSugestion.MinBlocksUntilInclusion = minBlocks
	feeSugestion.MinTimeUntilInclusion = float64(minBlocks) * avgBlockTime

	feeSugestion.MaxBlocksUntilInclusion = maxBlocks
	feeSugestion.MaxTimeUntilInclusion = float64(maxBlocks) * avgBlockTime
}
