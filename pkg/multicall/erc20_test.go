package multicall_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	"github.com/status-im/go-wallet-sdk/pkg/multicall"
)

func TestProcessERC20BalanceResult(t *testing.T) {
	word := func(v *big.Int) []byte {
		out := make([]byte, 32)
		v.FillBytes(out)
		return out
	}

	t.Run("a successful one-word answer is the balance", func(t *testing.T) {
		balance, err := multicall.ProcessERC20BalanceResult(multicall3.IMulticall3Result{Success: true, ReturnData: word(big.NewInt(12345))})
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(12345), balance)
	})

	t.Run("a successful zero word is a real zero", func(t *testing.T) {
		balance, err := multicall.ProcessERC20BalanceResult(multicall3.IMulticall3Result{Success: true, ReturnData: word(big.NewInt(0))})
		require.NoError(t, err)
		assert.Equal(t, int64(0), balance.Int64())
	})

	t.Run("a successful empty answer is an error, not a zero balance", func(t *testing.T) {
		// what Multicall3 returns for a target without code, e.g. an RPC upstream lacking the contract's state
		_, err := multicall.ProcessERC20BalanceResult(multicall3.IMulticall3Result{Success: true, ReturnData: []byte{}})
		assert.Error(t, err)
	})

	t.Run("a successful answer of the wrong width is an error", func(t *testing.T) {
		_, err := multicall.ProcessERC20BalanceResult(multicall3.IMulticall3Result{Success: true, ReturnData: []byte{1, 2, 3}})
		assert.Error(t, err)
	})

	t.Run("a failed call carries the revert data as the error", func(t *testing.T) {
		_, err := multicall.ProcessERC20BalanceResult(multicall3.IMulticall3Result{Success: false, ReturnData: []byte("revert")})
		require.Error(t, err)
		assert.Equal(t, "revert", err.Error())
	})
}

func TestBuildERC20BalanceCall_MatchesABIPack(t *testing.T) {
	erc20ABI, err := erc20.Erc20MetaData.GetAbi()
	require.NoError(t, err)
	for _, account := range []common.Address{
		{},
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0xfFfFfFfFfFfFfFfFfFfFfFfFfFfFfFfFfFfFfFfF"),
		common.HexToAddress("0x00000000000000000000000000000000000000a1"),
	} {
		token := common.HexToAddress("0x3333333333333333333333333333333333333333")
		expected, err := erc20ABI.Pack("balanceOf", account)
		require.NoError(t, err)
		call := multicall.BuildERC20BalanceCall(account, token)
		assert.Equal(t, expected, call.CallData, account.Hex())
		assert.Equal(t, token, call.Target)
	}
}

var callSink multicall3.IMulticall3Call

func TestBuildERC20BalanceCall_SingleAllocation(t *testing.T) {
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")
	token := common.HexToAddress("0x3333333333333333333333333333333333333333")
	allocs := testing.AllocsPerRun(100, func() {
		callSink = multicall.BuildERC20BalanceCall(account, token)
	})
	assert.Equal(t, float64(1), allocs)
}
