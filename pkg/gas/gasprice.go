package gas

import (
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

func suggestGasPrice(feeHistory *ethclient.FeeHistory, nBlocks int) *GasPrice {
	estimatedBaseFee := feeHistory.BaseFeePerGas[len(feeHistory.BaseFeePerGas)-1]

	ret := &GasPrice{
		BaseFeePerGas:           estimatedBaseFee,
		LowPriorityFeePerGas:    calculatePriorityFeeFromHistory(feeHistory, nBlocks, LowPriorityFeeIndex),
		MediumPriorityFeePerGas: calculatePriorityFeeFromHistory(feeHistory, nBlocks, MediumPriorityFeeIndex),
		HighPriorityFeePerGas:   calculatePriorityFeeFromHistory(feeHistory, nBlocks, HighPriorityFeeIndex),
	}

	return ret
}
