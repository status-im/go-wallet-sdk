package multicall3

import (
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

var (
	tokenA  = common.HexToAddress("0x00000000000000000000000000000000000000a1")
	tokenB  = common.HexToAddress("0x00000000000000000000000000000000000000b2")
	tokenC  = common.HexToAddress("0x00000000000000000000000000000000000000c3")
	account = common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
)

// balanceOf is the call data of balanceOf(owner).
func balanceOf(owner common.Address) []byte {
	return append([]byte{0x70, 0xa0, 0x82, 0x31}, common.LeftPadBytes(owner.Bytes(), 32)...)
}

// decode reads packed back with go-ethereum's decoder, which knows nothing of
// how packCalls lays the calls out.
func decode(t *testing.T, method string, packed []byte) (bool, []IMulticall3Call) {
	parsed, err := Multicall3MetaData.GetAbi()
	require.NoError(t, err)
	require.Equal(t, parsed.Methods[method].ID, packed[:4])
	args, err := parsed.Methods[method].Inputs.Unpack(packed[4:])
	require.NoError(t, err)
	raw := args[1].([]struct {
		Target   common.Address `json:"target"`
		CallData []byte         `json:"callData"`
	})
	calls := make([]IMulticall3Call, len(raw))
	for i, c := range raw {
		calls[i] = IMulticall3Call{Target: c.Target, CallData: c.CallData}
	}
	return args[0].(bool), calls
}

func pack(t *testing.T, method string, requireSuccess bool, calls []IMulticall3Call) []byte {
	parsed, err := Multicall3MetaData.GetAbi()
	require.NoError(t, err)
	return packCalls(parsed.Methods[method].ID, requireSuccess, calls)
}

func TestPackCalls_DecodesToTheCallsItWasGiven(t *testing.T) {
	tests := map[string][]IMulticall3Call{
		"none":             {},
		"one":              {{Target: tokenA, CallData: balanceOf(account)}},
		"same call data":   {{Target: tokenA, CallData: balanceOf(account)}, {Target: tokenB, CallData: balanceOf(account)}, {Target: tokenC, CallData: balanceOf(account)}},
		"each its own":     {{Target: tokenA, CallData: balanceOf(tokenA)}, {Target: tokenA, CallData: balanceOf(tokenB)}},
		"mixed and empty":  {{Target: tokenA, CallData: balanceOf(account)}, {Target: tokenB, CallData: []byte{}}, {Target: tokenC, CallData: balanceOf(account)}, {Target: tokenA, CallData: []byte{0xa3, 0xb1, 0xb3, 0x1d}}, {Target: tokenB, CallData: []byte{}}},
		"not a whole word": {{Target: tokenA, CallData: []byte{1, 2, 3}}, {Target: tokenB, CallData: make([]byte, 33)}},
	}
	for name, calls := range tests {
		t.Run(name, func(t *testing.T) {
			for _, method := range []string{"tryAggregate", "tryBlockAndAggregate"} {
				requireSuccess, decoded := decode(t, method, pack(t, method, true, calls))
				require.True(t, requireSuccess)
				require.Equal(t, calls, decoded)
			}
		})
	}
}

func TestPackCalls_KeepsOneCopyOfSharedCallData(t *testing.T) {
	const n = 1000
	calls := make([]IMulticall3Call, n)
	for i := range calls {
		calls[i] = IMulticall3Call{Target: common.BigToAddress(big.NewInt(int64(i + 1))), CallData: balanceOf(account)}
	}
	parsed, err := Multicall3MetaData.GetAbi()
	require.NoError(t, err)
	standard, err := parsed.Pack("tryBlockAndAggregate", false, calls)
	require.NoError(t, err)

	shared := pack(t, "tryBlockAndAggregate", false, calls)

	// A word each for the offset to the call, its target and the offset to its
	// call data, and the call data once.
	require.Len(t, shared, 4+3*abiWord+n*3*abiWord+3*abiWord)
	require.Less(t, len(shared)*100/len(standard), 51, "half the standard encoding")
}

func TestPackCalls_IsNoLargerThanTheStandardEncodingWithoutSharing(t *testing.T) {
	calls := []IMulticall3Call{{Target: tokenA, CallData: balanceOf(tokenA)}, {Target: tokenB, CallData: balanceOf(tokenB)}}
	parsed, err := Multicall3MetaData.GetAbi()
	require.NoError(t, err)
	standard, err := parsed.Pack("tryAggregate", true, calls)
	require.NoError(t, err)

	require.Len(t, pack(t, "tryAggregate", true, calls), len(standard))
}

// newMulticall3 runs the bytecode Multicall3 has on mainnet in a simulated chain.
func newMulticall3(t *testing.T, funded common.Address, balance *big.Int) (*Multicall3Caller, common.Address) {
	code, err := os.ReadFile("testdata/multicall3.runtime.hex")
	require.NoError(t, err)
	address := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
	backend := simulated.NewBackend(types.GenesisAlloc{
		address: {Code: common.FromHex(strings.TrimSpace(string(code)))},
		funded:  {Balance: balance},
	})
	t.Cleanup(func() { _ = backend.Close() })
	backend.Commit()
	caller, err := NewMulticall3Caller(address, backend.Client())
	require.NoError(t, err)
	return caller, address
}

func TestViewTryBlockAndAggregate_TheContractReadsSharedCallData(t *testing.T) {
	balance := big.NewInt(1_234_567)
	multicall, address := newMulticall3(t, account, balance)
	parsed, err := Multicall3MetaData.GetAbi()
	require.NoError(t, err)
	getEthBalance, err := parsed.Pack("getEthBalance", account)
	require.NoError(t, err)
	getBlockNumber, err := parsed.Pack("getBlockNumber")
	require.NoError(t, err)
	// The same call data in calls that are not next to each other, and a
	// target with no code, whose call succeeds with nothing to return.
	calls := []IMulticall3Call{
		{Target: address, CallData: getEthBalance},
		{Target: address, CallData: getBlockNumber},
		{Target: address, CallData: getEthBalance},
		{Target: tokenA, CallData: getEthBalance},
		{Target: address, CallData: getEthBalance},
	}

	blockNumber, _, results, err := multicall.ViewTryBlockAndAggregate(&bind.CallOpts{}, false, calls)

	require.NoError(t, err)
	require.Equal(t, int64(1), blockNumber.Int64())
	require.Len(t, results, len(calls))
	word := func(v *big.Int) []byte { return common.LeftPadBytes(v.Bytes(), 32) }
	for _, i := range []int{0, 2, 4} {
		require.True(t, results[i].Success)
		require.Equal(t, word(balance), results[i].ReturnData, "call %d", i)
	}
	require.Equal(t, word(blockNumber), results[1].ReturnData)
	require.True(t, results[3].Success)
	require.Empty(t, results[3].ReturnData)
}

func TestViewTryAggregate_TheContractReadsSharedCallData(t *testing.T) {
	balance := big.NewInt(42)
	multicall, address := newMulticall3(t, account, balance)
	parsed, err := Multicall3MetaData.GetAbi()
	require.NoError(t, err)
	getEthBalance, err := parsed.Pack("getEthBalance", account)
	require.NoError(t, err)
	calls := make([]IMulticall3Call, 300)
	for i := range calls {
		calls[i] = IMulticall3Call{Target: address, CallData: getEthBalance}
	}

	results, err := multicall.ViewTryAggregate(&bind.CallOpts{}, true, calls)

	require.NoError(t, err)
	require.Len(t, results, len(calls))
	for i, result := range results {
		require.True(t, result.Success, "call %d", i)
		require.Equal(t, common.LeftPadBytes(balance.Bytes(), 32), result.ReturnData, "call %d", i)
	}
}
