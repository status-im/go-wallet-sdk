package fetcher

//go:generate mockgen -destination=mock/fetcher.go . RPCClient,BatchCaller

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

type RPCClient interface {
	ChainID(ctx context.Context) (*big.Int, error)
	BatchCaller
	bind.ContractCaller
}

func FetchNativeBalances(
	ctx context.Context,
	addresses []common.Address,
	atBlock gethrpc.BlockNumber,
	rpcClient RPCClient,
	batchSize int,
) (BalancePerAccountAddress, error) {
	chainID, err := rpcClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	// For efficiency, try to use multicall first if available for this chain
	multicallAddress, exists := multicall3.GetMulticall3Address(chainID.Int64())
	if exists {
		multicallCaller, err := multicall3.NewMulticall3Caller(multicallAddress, rpcClient)
		if err == nil {
			return FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, multicallCaller, multicallAddress, batchSize)
		}
	}

	// As last resort, use less efficient batch call
	return FetchNativeBalancesStandard(ctx, addresses, atBlock, rpcClient, batchSize)
}

func FetchErc20Balances(
	ctx context.Context,
	addresses []common.Address,
	tokenAddresses []common.Address,
	atBlock gethrpc.BlockNumber,
	rpcClient RPCClient,
	batchSize int,
) (BalancePerAccountAndTokenAddress, error) {
	chainID, err := rpcClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	// For efficiency, try to use multicall first if available for this chain
	multicallAddress, exists := multicall3.GetMulticall3Address(chainID.Int64())
	if exists {
		multicallCaller, err := multicall3.NewMulticall3Caller(multicallAddress, rpcClient)
		if err == nil {
			return FetchErc20BalancesWithMulticall(ctx, addresses, tokenAddresses, atBlock, multicallCaller, batchSize)
		}
	}

	// As last resort, use less efficient batch call
	return FetchErc20BalancesStandard(ctx, addresses, tokenAddresses, atBlock, rpcClient, batchSize)
}
