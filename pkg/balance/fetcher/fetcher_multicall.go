package fetcher

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	"github.com/status-im/go-wallet-sdk/pkg/multicall"
)

func FetchNativeBalancesWithMulticall(
	ctx context.Context,
	accountAddresses []common.Address,
	atBlock gethrpc.BlockNumber,
	multicallCaller multicall.Caller,
	multicallAddress common.Address,
	batchSize int,
) (BalancePerAccountAddress, error) {
	balances := make(BalancePerAccountAddress)

	calls := make([]multicall3.IMulticall3Call, 0, len(accountAddresses))
	for _, accountAddress := range accountAddresses {
		calls = append(calls, multicall.BuildNativeBalanceCall(accountAddress, multicallAddress))
	}
	jobs := []multicall.Job{
		{
			Calls: calls,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return multicall.ProcessNativeBalanceResult(result)
			},
		},
	}

	results := multicall.RunSync(ctx, jobs, big.NewInt(int64(atBlock)), multicallCaller, batchSize)

	for _, result := range results {
		if result.Err != nil {
			return nil, result.Err
		}
		for i, accountAddress := range accountAddresses {
			callResult := result.Results[i]
			if callResult.Err != nil {
				continue // Skip failed individual calls
			}
			balance, ok := callResult.Value.(*big.Int)
			if !ok {
				continue
			}
			balances[accountAddress] = balance
		}
	}

	return balances, nil
}

func FetchErc20BalancesWithMulticall(
	ctx context.Context,
	accountAddresses []common.Address,
	tokenAddresses []common.Address,
	atBlock gethrpc.BlockNumber,
	multicallCaller multicall.Caller,
	batchSize int,
) (BalancePerAccountAndTokenAddress, error) {
	calls := make([]multicall3.IMulticall3Call, 0, len(accountAddresses)*len(tokenAddresses))
	balances := make(BalancePerAccountAndTokenAddress, len(accountAddresses))
	for _, accountAddress := range accountAddresses {
		balances[accountAddress] = make(BalancePerTokenAddress, len(tokenAddresses))
		for _, tokenAddress := range tokenAddresses {
			calls = append(calls, multicall.BuildERC20BalanceCall(accountAddress, tokenAddress))
		}
	}
	jobs := []multicall.Job{
		{
			Calls: calls,
			CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
				return multicall.ProcessERC20BalanceResult(result)
			},
		},
	}

	results := multicall.RunSync(ctx, jobs, big.NewInt(int64(atBlock)), multicallCaller, batchSize)

	for _, result := range results {
		if result.Err != nil {
			return nil, result.Err
		}
		idx := 0
		for _, accountAddress := range accountAddresses {
			for _, tokenAddress := range tokenAddresses {
				callResult := result.Results[idx]
				idx++
				if callResult.Err != nil {
					continue // Skip failed individual calls
				}
				balance, ok := callResult.Value.(*big.Int)
				if !ok {
					continue
				}
				balances[accountAddress][tokenAddress] = balance
			}
		}
	}

	return balances, nil
}
