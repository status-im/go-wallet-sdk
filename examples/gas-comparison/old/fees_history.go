package old

import (
	"context"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	ethCommon "github.com/ethereum/go-ethereum/common"
)

func isEIP1559Compatible(fh *ethereum.FeeHistory, chainID uint64) bool {
	// Since the Status Network is gasless chain, but EIP-1559 compatible, we should not rely on checking the BaseFeePerGas, that's why we have this special case.
	eip1559Enabled, err := IsPartiallyOrFullyGaslessChainEIP1559Compatible(chainID)
	if err == nil {
		return eip1559Enabled
	}

	if len(fh.BaseFee) == 0 {
		return false
	}

	for _, fee := range fh.BaseFee {
		if fee.Cmp(big.NewInt(0)) != 0 {
			return true
		}
	}

	return false
}

func (f *FeeManager) getFeeHistory(ctx context.Context, chainID uint64, newestBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
	blockCount := uint64(10) // use the last 10 blocks for L1 chains
	if chainID != EthereumMainnet &&
		chainID != EthereumSepolia &&
		chainID != AnvilMainnet {
		blockCount = 50 // use the last 50 blocks for L2 chains
	}

	return f.ethClient.FeeHistory(ctx, blockCount, newestBlock, rewardPercentiles)
}

func (f *FeeManager) getGaslessParamsForAccount(ctx context.Context, chainID uint64, address ethCommon.Address) (baseFee *big.Int, priorityFee *big.Int, err error) {
	if !IsPartiallyOrFullyGaslessChain(chainID) {
		return nil, nil, nil
	}

	toAddress := ZeroAddress()
	msg := ethereum.CallMsg{
		From:  address,
		To:    &toAddress,
		Value: ZeroBigIntValue(),
	}

	result, err := f.ethClient.LineaEstimateGas(ctx, msg)
	if err != nil {
		return
	}

	baseFee = result.BaseFeePerGas
	priorityFee = result.PriorityFeePerGas

	return
}
