package gas

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
)

func estimateTxGas(ctx context.Context, ethClient EthClient, callMsg *ethereum.CallMsg) (*big.Int, error) {
	gasUsed, err := ethClient.EstimateGas(ctx, *callMsg)
	if err != nil {
		return nil, err
	}
	return big.NewInt(0).SetUint64(gasUsed), nil
}
