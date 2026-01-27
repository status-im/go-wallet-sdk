package gas

import (
	"context"
	"math/big"
)

func estimateLegacyInclusion(ctx context.Context, blockNumberClient blockNumberReader, gasPrice *big.Int, nBlocks int, avgBlockTime float64) Inclusion {
	if gasPrice == nil || gasPrice.Sign() <= 0 {
		return unknownInclusion(avgBlockTime)
	}

	sortedGasPrices, err := tryGetSortedLegacyGasPrices(ctx, blockNumberClient, nBlocks)
	if err != nil || len(sortedGasPrices) == 0 {
		return unknownInclusion(avgBlockTime)
	}

	return estimateLegacyInclusionFromSorted(gasPrice, sortedGasPrices, avgBlockTime)
}

func estimateLegacyInclusionFromSorted(gasPrice *big.Int, sortedGasPrices []*big.Int, avgBlockTime float64) Inclusion {
	minBlocks, maxBlocks := estimateBlocksUntilInclusionLegacy(gasPrice, sortedGasPrices)
	return Inclusion{
		MinBlocksUntilInclusion: minBlocks,
		MaxBlocksUntilInclusion: maxBlocks,
		MinTimeUntilInclusion:   float64(minBlocks) * avgBlockTime,
		MaxTimeUntilInclusion:   maxTimeFromBlocks(maxBlocks, avgBlockTime),
	}
}

func unknownInclusion(avgBlockTime float64) Inclusion {
	return Inclusion{
		MinBlocksUntilInclusion: 1,
		MaxBlocksUntilInclusion: -1,
		MinTimeUntilInclusion:   avgBlockTime,
		MaxTimeUntilInclusion:   -1,
	}
}

func maxTimeFromBlocks(maxBlocks int, avgBlockTime float64) float64 {
	if maxBlocks < 0 {
		return -1
	}
	return float64(maxBlocks) * avgBlockTime
}

// estimateBlocksUntilInclusionLegacy estimates inclusion for legacy chains using a simple percentile-based model
// over recent block tx gas prices.
// returned value for maxBlocks is -1 if the upper bound is unknown
func estimateBlocksUntilInclusionLegacy(gasPrice *big.Int, sortedGasPrices []*big.Int) (minBlocks int, maxBlocks int) {
	inclusions := []struct {
		inclusionInBlock int
		percentile       float64
	}{
		{2, baseFeePercentileSecondBlock},
		{3, baseFeePercentileThirdBlock},
		{4, baseFeePercentileFourthBlock},
		{5, baseFeePercentileFifthBlock},
		{6, baseFeePercentileSixthBlock},
	}

	inclusionIdx := -1
	for idx, p := range inclusions {
		threshold := getPercentile(sortedGasPrices, p.percentile)
		if gasPrice.Cmp(threshold) >= 0 {
			inclusionIdx = idx
			break
		}
	}

	minBlocks = 1
	maxBlocks = inclusions[len(inclusions)-1].inclusionInBlock + 1

	if inclusionIdx < 0 {
		minBlocks = inclusions[len(inclusions)-1].inclusionInBlock
		return
	}
	if inclusionIdx == 0 {
		maxBlocks = inclusions[0].inclusionInBlock
		return
	}

	minBlocks = inclusions[inclusionIdx-1].inclusionInBlock
	maxBlocks = inclusions[inclusionIdx].inclusionInBlock
	return
}
