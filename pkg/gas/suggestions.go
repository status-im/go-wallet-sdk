package gas

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// DefaultConfig returns default configuration
func DefaultConfig(chainClass ChainClass) SuggestionsConfig {
	switch chainClass {
	case ChainClassL1:
		return SuggestionsConfig{
			NetworkCongestionBlocks:           10,
			GasPriceEstimationBlocks:          10,
			LowRewardPercentile:               10,
			MediumRewardPercentile:            45,
			HighRewardPercentile:              90,
			LowBaseFeeMultiplier:              1.025, // 2.5% buffer for base fee
			MediumBaseFeeMultiplier:           1.025,
			HighBaseFeeMultiplier:             1.025,
			LowBaseFeeCongestionMultiplier:    0.0, // No congestion-based adjustment for Low level
			MediumBaseFeeCongestionMultiplier: 10.0,
			HighBaseFeeCongestionMultiplier:   10.0,

			LowGasPriceMultiplier:    1.00,
			MediumGasPriceMultiplier: 1.10,
			HighGasPriceMultiplier:   1.20,
		}
	}

	return SuggestionsConfig{
		NetworkCongestionBlocks:           10,
		GasPriceEstimationBlocks:          50,
		LowRewardPercentile:               10,
		MediumRewardPercentile:            45,
		HighRewardPercentile:              90,
		LowBaseFeeMultiplier:              1.025,
		MediumBaseFeeMultiplier:           4.1,
		HighBaseFeeMultiplier:             10.25,
		LowBaseFeeCongestionMultiplier:    0.0, // No congestion-based adjustment at any level
		MediumBaseFeeCongestionMultiplier: 0.0,
		HighBaseFeeCongestionMultiplier:   0.0,

		LowGasPriceMultiplier:    1.00,
		MediumGasPriceMultiplier: 1.10,
		HighGasPriceMultiplier:   1.20,
	}
}

func GetChainSuggestions(ctx context.Context, gasClient GasClient, params ChainParameters, config SuggestionsConfig, account common.Address) (suggestions *FeeSuggestions, err error) {
	if params.FeeModel == FeeModelLegacy {
		suggestions, err = getLegacyChainSuggestions(ctx, gasClient, params, config)
		if err != nil {
			return
		}
		suggestions.FeeModel = FeeModelLegacy
		return
	}

	switch params.ChainClass {
	case ChainClassL1:
		suggestions, err = getL1ChainSuggestions(ctx, gasClient, params, config)
	case ChainClassLineaStack:
		suggestions, err = getLineaChainSuggestions(ctx, gasClient, params, config, account)
	default:
		suggestions, err = getL2ChainSuggestions(ctx, gasClient, params, config)
	}
	if err != nil {
		return
	}
	suggestions.FeeModel = FeeModelEIP1559
	return
}

func GetTxSuggestions(ctx context.Context, gasClient GasClient, params ChainParameters, config SuggestionsConfig, callMsg *ethereum.CallMsg) (suggestions *TxSuggestions, err error) {
	if callMsg == nil {
		return nil, fmt.Errorf("call msg is required for tx suggestions")
	}

	if params.FeeModel == FeeModelLegacy {
		suggestions, err = getLegacyTxSuggestions(ctx, gasClient, params, config, callMsg)
		if err != nil {
			return
		}
		suggestions.FeeSuggestions.FeeModel = FeeModelLegacy
		return
	}

	switch params.ChainClass {
	case ChainClassL1:
		suggestions, err = getL1TxSuggestions(ctx, gasClient, params, config, callMsg)
	case ChainClassLineaStack:
		suggestions, err = getLineaTxSuggestions(ctx, gasClient, params, config, callMsg)
	default:
		suggestions, err = getL2TxSuggestions(ctx, gasClient, params, config, callMsg)
	}
	if err != nil {
		return
	}
	suggestions.FeeSuggestions.FeeModel = FeeModelEIP1559
	return
}

func EstimateInclusion(ctx context.Context, inclusionEstimatorClient BlockInclusionEstimator, params ChainParameters, config SuggestionsConfig, fee Fee) (*Inclusion, error) {
	if fee.GasPrice != nil {
		inclusion := estimateLegacyInclusion(ctx, inclusionEstimatorClient, fee.GasPrice, config.GasPriceEstimationBlocks, params.NetworkBlockTime)
		return &inclusion, nil
	}

	blockCount := uint64(max(config.GasPriceEstimationBlocks, config.NetworkCongestionBlocks))
	rewardPercentiles := []float64{config.MediumRewardPercentile}

	feeHistory, err := getFeeHistory(ctx, inclusionEstimatorClient, blockCount, nil, rewardPercentiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get fee history: %w", err)
	}

	sortedBaseFees := getSortedBaseFees(feeHistory)
	sortedMediumPriorityFees := getSortedPriorityFees(feeHistory, MediumPriorityFeeIndex)

	inclusion := estimateInclusion(fee, sortedBaseFees, sortedMediumPriorityFees, params.NetworkBlockTime)
	return &inclusion, nil
}

func ResolveFeeModel(ctx context.Context, feeResolverClient FeeModelResolver) (FeeModel, error) {
	// Some providers return error if the endpoint is not supported for the chain, while others return some value or zeroed value.
	// The code below is reliably trying to detect if the endpoint is supported for the chain.
	// If SuggestGasTipCap is not supported, means the chain is legacy, not EIP-1559 aligned.
	var (
		maxPriorityFeePerGas *big.Int
		feeHistory           *ethereum.FeeHistory
		err                  error
	)
	if maxPriorityFeePerGas, err = feeResolverClient.SuggestGasTipCap(ctx); err != nil {
		return FeeModelLegacy, nil
	}

	// If FeeHistory is not supported, means the chain is legacy, not EIP-1559 aligned.
	var rewardPercentiles = []float64{50.0} // any value from 0 to 100 is valid in this context
	if feeHistory, err = getFeeHistory(ctx, feeResolverClient, 1, nil, rewardPercentiles); err != nil {
		return FeeModelLegacy, nil
	}

	// If it's not present, means the chain is legacy, not EIP-1559 aligned.
	if fm, ok := detectFeeModelFromLatestBlock(ctx, feeResolverClient); ok {
		if fm == FeeModelLegacy {
			return FeeModelLegacy, nil
		}

		// the chain is not EIP-1559 aligned if maxPriorityFeePerGas is zero and baseFee and reward are zero in feeHistory
		isLegacy := maxPriorityFeePerGas.Sign() == 0
		if !isLegacy {
			return FeeModelEIP1559, nil
		}

		// check if all base fees are zero
		for _, baseFee := range feeHistory.BaseFee {
			if baseFee.Sign() != 0 {
				return FeeModelEIP1559, nil
			}
		}

		// check if all rewards are zero
		for _, rewards := range feeHistory.Reward {
			for _, reward := range rewards {
				if reward.Sign() != 0 {
					return FeeModelEIP1559, nil
				}
			}
		}

		return FeeModelLegacy, nil
	}

	return "", fmt.Errorf("unable to detect fee model")
}

func detectFeeModelFromLatestBlock(ctx context.Context, feeResolverClient FeeModelResolver) (FeeModel, bool) {
	block, err := feeResolverClient.EthGetBlockByNumberWithFullTxs(ctx, nil)
	if err != nil || block == nil {
		// If BlockByNumber failed, try to get it using the block number.
		n, err1 := feeResolverClient.BlockNumber(ctx)
		if err1 == nil {
			block, err = feeResolverClient.EthGetBlockByNumberWithFullTxs(ctx, new(big.Int).SetUint64(n))
		}
	}
	if err != nil || block == nil {
		return "", false
	}

	if block.BaseFeePerGas == nil {
		return FeeModelLegacy, true
	}
	return FeeModelEIP1559, true
}
