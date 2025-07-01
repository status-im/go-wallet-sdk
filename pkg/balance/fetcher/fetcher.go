package fetcher

//go:generate mockgen -destination=mock/fetcher.go . RPCClient,BatchCaller,BalanceScanner

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
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

	// For efficiency, try to use balancescanner first if available for this chain
	if balanceScanner := getBalanceScanner(chainID.Uint64(), atBlock, rpcClient); balanceScanner != nil {
		return FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, balanceScanner, batchSize)
	}

	// As last resort, use less efficient batch call
	return FetchNativeBalancesStandard(ctx, addresses, atBlock, rpcClient, batchSize)
}
