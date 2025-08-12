package gas

import (
	"context"
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

func getLineaTxSuggestions(ctx context.Context, ethClient EthClient, config SuggestionsConfig, callMsg *ethereum.CallMsg) (*TxSuggestions, error) {
	lineaEstimateGasResult, err := estimateLineaTxGas(ctx, ethClient, callMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate linea gas: %w", err)
	}

	gasPrice, err := suggestLineaGasPrice(lineaEstimateGasResult)
	if err != nil {
		return nil, fmt.Errorf("failed to get linea suggestions: %w", err)
	}

	twiceBaseFee := big.NewInt(0).Mul(gasPrice.BaseFeePerGas, big.NewInt(2))

	ret := &TxSuggestions{
		GasLimit: lineaEstimateGasResult.GasLimit,
		FeeSuggestions: &FeeSuggestions{
			EstimatedBaseFee:      gasPrice.BaseFeePerGas,
			NetworkCongestion:     0,
			PriorityFeeLowerBound: gasPrice.LowPriorityFeePerGas,
			PriorityFeeUpperBound: gasPrice.HighPriorityFeePerGas,
			Low: FeeSuggestion{
				MaxPriorityFeePerGas: gasPrice.LowPriorityFeePerGas,
				MaxFeePerGas:         big.NewInt(0).Add(twiceBaseFee, gasPrice.LowPriorityFeePerGas),
			},
			Medium: FeeSuggestion{
				MaxPriorityFeePerGas: gasPrice.MediumPriorityFeePerGas,
				MaxFeePerGas:         big.NewInt(0).Add(twiceBaseFee, gasPrice.MediumPriorityFeePerGas),
			},
			High: FeeSuggestion{
				MaxPriorityFeePerGas: gasPrice.HighPriorityFeePerGas,
				MaxFeePerGas:         big.NewInt(0).Add(twiceBaseFee, gasPrice.HighPriorityFeePerGas),
			},
		},
	}

	// Calculate inclusions
	blockCount := uint64(config.NetworkCongestionBlocks)
	rewardPercentiles := []float64{config.MediumRewardPercentile}

	feeHistory, err := ethClient.FeeHistory(ctx, blockCount, nil, rewardPercentiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get fee history: %w", err)
	}

	fillInclusionsLinea(ret, feeHistory, config.NetworkBlockTime)

	return ret, nil
}

func fillInclusionsLinea(txSuggestions *TxSuggestions, feeHistory *ethclient.FeeHistory, avgBlockTime float64) {
	sortedBaseFees := make([]*big.Int, len(feeHistory.BaseFeePerGas[:len(feeHistory.BaseFeePerGas)-2]))
	copy(sortedBaseFees, feeHistory.BaseFeePerGas[:len(feeHistory.BaseFeePerGas)-2])
	slices.SortFunc(sortedBaseFees, func(a, b *big.Int) int {
		return a.Cmp(b)
	})

	sortedMediumPriorityFees := extractPriorityFeesFromHistory(feeHistory, 0)
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
