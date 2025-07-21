package gas

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
)

func getL1Suggestions(ctx context.Context, ethClient EthClient, config SuggestionsConfig, callMsg *ethereum.CallMsg) (*TxSuggestions, error) {
	ret := &TxSuggestions{}

	gasLimit, err := estimateTxGas(ctx, ethClient, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	ret.GasLimit = gasLimit

	blockCount := uint64(max(config.GasPriceEstimationBlocks, config.NetworkCongestionBlocks))
	rewardPercentiles := []float64{config.LowRewardPercentile, config.MediumRewardPercentile, config.HighRewardPercentile}

	feeHistory, err := ethClient.FeeHistory(ctx, blockCount, nil, rewardPercentiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get fee history: %w", err)
	}

	gasPrice := suggestGasPrice(feeHistory, config.GasPriceEstimationBlocks)

	// Apply multiplier to BaseFee
	suggestedBaseFee := new(big.Int).Mul(gasPrice.BaseFeePerGas, big.NewInt(int64(config.BaseFeeMultiplier*1000)))
	suggestedBaseFee.Div(suggestedBaseFee, big.NewInt(1000))

	// Use congestion-based logic
	networkCongestion := calculateNetworkCongestionFromHistory(feeHistory, config.NetworkCongestionBlocks)
	congestionFactor := new(big.Float).SetFloat64(10 * networkCongestion)

	baseFeeFloat := new(big.Float).SetInt(suggestedBaseFee)
	baseFeeFloat.Mul(baseFeeFloat, congestionFactor)
	additionBasedOnCongestion := new(big.Int)
	baseFeeFloat.Int(additionBasedOnCongestion)

	lowBaseFee := new(big.Int).Set(suggestedBaseFee)

	mediumBaseFee := new(big.Int).Add(suggestedBaseFee, additionBasedOnCongestion)

	highBaseFee := new(big.Int).Mul(suggestedBaseFee, big.NewInt(2))
	highBaseFee.Add(highBaseFee, additionBasedOnCongestion)

	ret.FeeSuggestions = &FeeSuggestions{
		EstimatedBaseFee:      gasPrice.BaseFeePerGas,
		NetworkCongestion:     networkCongestion,
		PriorityFeeLowerBound: gasPrice.LowPriorityFeePerGas,
		PriorityFeeUpperBound: gasPrice.HighPriorityFeePerGas,
		Low: FeeSuggestion{
			MaxPriorityFeePerGas: gasPrice.LowPriorityFeePerGas,
			MaxFeePerGas:         big.NewInt(0).Add(lowBaseFee, gasPrice.LowPriorityFeePerGas),
		},
		Medium: FeeSuggestion{
			MaxPriorityFeePerGas: gasPrice.MediumPriorityFeePerGas,
			MaxFeePerGas:         big.NewInt(0).Add(mediumBaseFee, gasPrice.MediumPriorityFeePerGas),
		},
		High: FeeSuggestion{
			MaxPriorityFeePerGas: gasPrice.HighPriorityFeePerGas,
			MaxFeePerGas:         big.NewInt(0).Add(highBaseFee, gasPrice.HighPriorityFeePerGas),
		},
	}

	// Calculate inclusions
	fillInclusions(ret, feeHistory, config.NetworkBlockTime)

	return ret, nil
}
