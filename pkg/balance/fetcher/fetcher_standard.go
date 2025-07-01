package fetcher

import (
	"context"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
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
