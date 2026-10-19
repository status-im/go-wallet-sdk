package ethclient

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// CallContractRaw executes eth_call like CallContract, with the call argument
// already encoded as JSON (e.g. {"from":"0x…","input":"0x…","to":"0x…"}). It
// spares re-encoding large calldata, which CallContract does on every call.
func (c *Client) CallContractRaw(ctx context.Context, callArg json.RawMessage, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	if err := c.rpcClient.CallContext(ctx, &result, "eth_call", callArg, toBlockNumArg(blockNumber)); err != nil {
		return nil, err
	}
	return result, nil
}
