package gas

//go:generate mockgen -destination=mock/ethclient.go . GasClient,FeeModelResolver,BlockInclusionEstimator

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

type feeHistoryReader interface {
	FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error)
}

type blockNumberReader interface {
	BlockNumber(ctx context.Context) (uint64, error)
	EthGetBlockByNumberWithFullTxs(ctx context.Context, number *big.Int) (*ethclient.BlockWithFullTxs, error)
}

type gasPriceReader interface {
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
}

type BlockInclusionEstimator interface {
	feeHistoryReader
	blockNumberReader
}

type FeeModelResolver interface {
	feeHistoryReader
	blockNumberReader
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
}

type GasClient interface {
	feeHistoryReader
	blockNumberReader
	gasPriceReader
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	LineaEstimateGas(ctx context.Context, msg ethereum.CallMsg) (*ethclient.LineaEstimateGasResult, error)
}
