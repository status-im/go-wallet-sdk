package fetcher

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/balancescanner"
)

// Returns balance scanner instance if it's available for this chain and was deployed at the given block number
func getBalanceScanner(chainID uint64, atBlock gethrpc.BlockNumber, backend bind.ContractCaller) BalanceScanner {
	balanceScannerData := balancescanner.GetContractData(chainID)
	if balanceScannerData != nil && (atBlock < 0 || atBlock >= gethrpc.BlockNumber(balanceScannerData.CreatedAtBlock)) { //nolint:gosec
		balanceScanner, err := balancescanner.NewBalancescannerCaller(balanceScannerData.Address, backend)
		if err != nil {
			return nil
		}
		return balanceScanner
	}
	return nil
}
