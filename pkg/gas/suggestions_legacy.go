package gas

import (
	"context"
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum"
)

const (
	legacyMaxBlockSample = 10
	legacyMaxTxSample    = 1000
)

func getLegacyChainSuggestions(ctx context.Context, gasClient GasClient, params ChainParameters, config SuggestionsConfig) (*FeeSuggestions, error) {
	txSuggestions, err := getLegacyTxSuggestions(ctx, gasClient, params, config, nil)
	if err != nil {
		return nil, err
	}
	return txSuggestions.FeeSuggestions, nil
}

func getLegacyTxSuggestions(ctx context.Context, gasClient GasClient, params ChainParameters, config SuggestionsConfig, callMsg *ethereum.CallMsg) (*TxSuggestions, error) {
	ret := &TxSuggestions{
		GasLimit: big.NewInt(0),
	}

	if callMsg != nil {
		gasLimit, err := gasClient.EstimateGas(ctx, *callMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
		ret.GasLimit = big.NewInt(0).SetUint64(gasLimit)
	}

	gasPrice, err := gasClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %w", err)
	}

	sortedGasPrices, _ := tryGetSortedLegacyGasPrices(ctx, gasClient, config.GasPriceEstimationBlocks)

	var low, medium, high *big.Int
	if len(sortedGasPrices) > 0 {
		low = getPercentile(sortedGasPrices, config.LowRewardPercentile)
		medium = getPercentile(sortedGasPrices, config.MediumRewardPercentile)
		high = getPercentile(sortedGasPrices, config.HighRewardPercentile)
	} else {
		low = mulBigIntFloat(gasPrice, config.LowGasPriceMultiplier)
		medium = mulBigIntFloat(gasPrice, config.MediumGasPriceMultiplier)
		high = mulBigIntFloat(gasPrice, config.HighGasPriceMultiplier)
	}

	ret.FeeSuggestions = &FeeSuggestions{
		GasPrice:          gasPrice,
		NetworkCongestion: 0,
		Low: Fee{
			GasPrice: low,
		},
		Medium: Fee{
			GasPrice: medium,
		},
		High: Fee{
			GasPrice: high,
		},
	}

	if len(sortedGasPrices) > 0 {
		ret.FeeSuggestions.LowInclusion = estimateLegacyInclusionFromSorted(low, sortedGasPrices, params.NetworkBlockTime)
		ret.FeeSuggestions.MediumInclusion = estimateLegacyInclusionFromSorted(medium, sortedGasPrices, params.NetworkBlockTime)
		ret.FeeSuggestions.HighInclusion = estimateLegacyInclusionFromSorted(high, sortedGasPrices, params.NetworkBlockTime)
	} else {
		ret.FeeSuggestions.LowInclusion = unknownInclusion(params.NetworkBlockTime)
		ret.FeeSuggestions.MediumInclusion = unknownInclusion(params.NetworkBlockTime)
		ret.FeeSuggestions.HighInclusion = unknownInclusion(params.NetworkBlockTime)
	}

	return ret, nil
}

func tryGetSortedLegacyGasPrices(ctx context.Context, blockNumberClient blockNumberReader, nBlocks int) ([]*big.Int, error) {
	latest, err := blockNumberClient.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	blocks := min(max(nBlocks, 1), legacyMaxBlockSample)

	out := make([]*big.Int, 0, min(legacyMaxTxSample, blocks*50))
	for i := 0; i < blocks && len(out) < legacyMaxTxSample; i++ {
		n := int64(latest) - int64(i)
		if n < 0 {
			break
		}
		b, err := blockNumberClient.EthGetBlockByNumberWithFullTxs(ctx, big.NewInt(n))
		if err != nil || b == nil {
			continue
		}
		for _, tx := range b.Transactions {
			if tx.GasPrice == nil || tx.GasPrice.Sign() <= 0 {
				continue
			}
			out = append(out, new(big.Int).Set(tx.GasPrice))
			if len(out) >= legacyMaxTxSample {
				break
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no legacy gas prices sampled")
	}

	slices.SortFunc(out, func(a, b *big.Int) int {
		return a.Cmp(b)
	})
	return out, nil
}

func mulBigIntFloat(v *big.Int, mul float64) *big.Int {
	if v == nil {
		return big.NewInt(0)
	}
	f := new(big.Float).SetInt(v)
	f.Mul(f, big.NewFloat(mul))
	out := new(big.Int)
	f.Int(out)
	return out
}
