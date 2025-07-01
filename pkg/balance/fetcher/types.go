package fetcher

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BalancePerAccountAddress = map[common.Address]*big.Int

type BalancePerTokenAddress = map[common.Address]*big.Int

type BalancePerAccountAndTokenAddress = map[common.Address]BalancePerTokenAddress
