package data

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

type GasData struct {
	LatestBlock          *ethclient.BlockWithFullTxs `json:"latestBlock"`
	FeeHistory           *ethereum.FeeHistory        `json:"feeHistory"`
	GasPrice             *big.Int                    `json:"gasPrice"`
	MaxPriorityFeePerGas *big.Int                    `json:"maxPriorityFeePerGas"`
	InfuraSuggestedFees  *infura.GasResponse         `json:"infuraSuggestedFees"`
}
