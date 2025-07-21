package gas

import (
	"context"

	"github.com/ethereum/go-ethereum"
)

// DefaultConfig returns default configuration
func DefaultConfig() SuggestionsConfig {
	return SuggestionsConfig{
		ChainClass:               ChainClassL1,
		NetworkBlockTime:         12,
		NetworkCongestionBlocks:  10,
		GasPriceEstimationBlocks: 10,
		LowRewardPercentile:      10,
		MediumRewardPercentile:   45,
		HighRewardPercentile:     90,
		BaseFeeMultiplier:        1.025, // 2.5% buffer for base fee
	}
}

func GetTxSuggestions(ctx context.Context, ethClient EthClient, config SuggestionsConfig, callMsg *ethereum.CallMsg) (*TxSuggestions, error) {
	switch config.ChainClass {
	case ChainClassL1:
		return getL1Suggestions(ctx, ethClient, config, callMsg)
	case ChainClassLineaStack:
		return getLineaTxSuggestions(ctx, ethClient, config, callMsg)
	}
	return getL2Suggestions(ctx, ethClient, config, callMsg)
}
