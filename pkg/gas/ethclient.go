package gas

//go:generate mockgen -destination=mock/ethclient.go . EthClient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

type EthClient interface {
	FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethclient.FeeHistory, error)
	BlockNumber(ctx context.Context) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethclient.BlockWithFullTxs, error)
	LineaEstimateGas(ctx context.Context, msg ethereum.CallMsg) (*ethclient.LineaEstimateGasResult, error)
}
