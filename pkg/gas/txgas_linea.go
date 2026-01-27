package gas

import (
	"context"

	"github.com/ethereum/go-ethereum"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

func estimateLineaTxGas(ctx context.Context, gasClient GasClient, callMsg *ethereum.CallMsg) (*ethclient.LineaEstimateGasResult, error) {
	return gasClient.LineaEstimateGas(ctx, *callMsg)
}
