package fetcher

import (
	"context"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/balancescanner"
)

type BalanceScanner interface {
	EtherBalances(opts *bind.CallOpts, addresses []common.Address) ([]balancescanner.BalanceScannerResult, error)
	TokenBalances(opts *bind.CallOpts, addresses []common.Address, tokenAddress common.Address) ([]balancescanner.BalanceScannerResult, error)
	TokensBalance(opts *bind.CallOpts, owner common.Address, contracts []common.Address) ([]balancescanner.BalanceScannerResult, error)
}

func FetchNativeBalancesWithBalanceScanner(
	ctx context.Context,
	accountAddresses []common.Address,
	atBlock gethrpc.BlockNumber,
	balanceScanner BalanceScanner,
	batchSize int,
) (map[common.Address]*big.Int, error) {
	balances := make(BalancePerAccountAddress)

	for chunk := range slices.Chunk(accountAddresses, batchSize) {
		res, err := balanceScanner.EtherBalances(&bind.CallOpts{
			Context:     ctx,
			BlockNumber: big.NewInt(int64(atBlock)),
		}, chunk)
		if err != nil {
			return nil, err
		}
		for idx, account := range chunk {
			balance := new(big.Int)
			balance.SetBytes(res[idx].Data)
			balances[account] = balance
		}
	}

	return balances, nil
}

func parseBalanceBytes(data []byte) *big.Int {
	balance := new(big.Int)
	balance.SetBytes(data)
	return balance
}

func FetchErc20BalancesWithBalanceScanner(
	ctx context.Context,
	accountAddresses []common.Address,
	tokenAddresses []common.Address,
	atBlock gethrpc.BlockNumber,
	balanceScanner BalanceScanner,
	batchSize int,
) (BalancePerAccountAndTokenAddress, error) {
	balances := make(BalancePerAccountAndTokenAddress, len(accountAddresses))
	for _, accountAddress := range accountAddresses {
		balances[accountAddress] = make(BalancePerTokenAddress, len(tokenAddresses))
	}

	if len(accountAddresses) > len(tokenAddresses) {
		// Loop over tokens, batch accounts
		for _, tokenAddress := range tokenAddresses {
			for chunk := range slices.Chunk(accountAddresses, batchSize) {
				res, err := balanceScanner.TokenBalances(&bind.CallOpts{
					Context:     ctx,
					BlockNumber: big.NewInt(int64(atBlock)),
				}, chunk, tokenAddress)
				if err != nil {
					return nil, err
				}
				for idx, account := range chunk {
					balances[account][tokenAddress] = parseBalanceBytes(res[idx].Data)
				}
			}
		}
	} else {
		// Loop over accounts, batch tokens
		for _, accountAddress := range accountAddresses {
			for chunk := range slices.Chunk(tokenAddresses, batchSize) {
				res, err := balanceScanner.TokensBalance(&bind.CallOpts{
					Context:     ctx,
					BlockNumber: big.NewInt(int64(atBlock)),
				}, accountAddress, chunk)
				if err != nil {
					return nil, err
				}
				for idx, token := range chunk {
					balances[accountAddress][token] = parseBalanceBytes(res[idx].Data)
				}
			}
		}
	}

	return balances, nil
}
