package fetcher

import (
	"context"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
)

type BatchCaller interface {
	BatchCallContext(ctx context.Context, b []gethrpc.BatchElem) error
}

func FetchNativeBalancesStandard(
	ctx context.Context,
	accountAddresses []common.Address,
	atBlock gethrpc.BlockNumber,
	batchCaller BatchCaller,
	batchSize int,
) (map[common.Address]*big.Int, error) {
	balances := make(BalancePerAccountAddress)

	for chunk := range slices.Chunk(accountAddresses, batchSize) {
		batch := make([]gethrpc.BatchElem, len(chunk))
		for i, address := range chunk {
			res := (*hexutil.Big)(big.NewInt(0))
			batch[i] = gethrpc.BatchElem{
				Method: "eth_getBalance",
				Args:   []interface{}{address, atBlock},
				Result: res,
			}
		}
		err := batchCaller.BatchCallContext(ctx, batch)
		if err != nil {
			return nil, err
		}

		for i, elem := range batch {
			if elem.Error != nil {
				return nil, elem.Error
			}
			balances[chunk[i]] = elem.Result.(*hexutil.Big).ToInt()
		}
	}

	return balances, nil
}

func FetchErc20BalancesStandard(
	ctx context.Context,
	accountAddresses []common.Address,
	tokenAddresses []common.Address,
	atBlock gethrpc.BlockNumber,
	batchCaller BatchCaller,
	batchSize int,
) (BalancePerAccountAndTokenAddress, error) {
	balances := make(BalancePerAccountAndTokenAddress, len(accountAddresses))
	for _, accountAddress := range accountAddresses {
		balances[accountAddress] = make(BalancePerTokenAddress, len(tokenAddresses))
	}

	type pair struct {
		accountAddress common.Address
		tokenAddress   common.Address
	}

	pairs := make([]pair, 0, len(accountAddresses)*len(tokenAddresses))
	for _, accountAddress := range accountAddresses {
		for _, tokenAddress := range tokenAddresses {
			pairs = append(pairs, pair{accountAddress, tokenAddress})
		}
	}

	abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	for chunk := range slices.Chunk(pairs, batchSize) {
		batch := make([]gethrpc.BatchElem, len(chunk))
		for i, pair := range chunk {
			res := (*hexutil.Big)(big.NewInt(0))

			input, err := abi.Pack("balanceOf", pair.accountAddress)
			if err != nil {
				return nil, err
			}
			batch[i] = gethrpc.BatchElem{
				Method: "eth_call",
				Args:   []interface{}{map[string]interface{}{"to": pair.tokenAddress, "data": input, "blockNumber": atBlock}, pair.tokenAddress},
				Result: res,
			}
		}

		err := batchCaller.BatchCallContext(ctx, batch)
		if err != nil {
			return nil, err
		}

		for i, elem := range batch {
			if elem.Error != nil {
				return nil, elem.Error
			}
			balances[chunk[i].accountAddress][chunk[i].tokenAddress] = elem.Result.(*hexutil.Big).ToInt()
		}
	}
	return balances, nil
}
