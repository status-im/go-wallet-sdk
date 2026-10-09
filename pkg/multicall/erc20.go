package multicall

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

// erc20BalanceOfSelector is the selector of "balanceOf(address)".
var erc20BalanceOfSelector = [4]byte{0x70, 0xa0, 0x82, 0x31}

// Call for ERC20 function "balanceOf(owner)". The calldata is encoded by hand,
// byte-identical to abi.Pack: a fetch builds one call per token and account.
func BuildERC20BalanceCall(accountAddress common.Address, tokenAddress common.Address) multicall3.IMulticall3Call {
	callData := make([]byte, 4+32)
	copy(callData, erc20BalanceOfSelector[:])
	copy(callData[4+32-common.AddressLength:], accountAddress[:])

	return multicall3.IMulticall3Call{
		Target:   tokenAddress,
		CallData: callData,
	}
}

// balanceOf(address) returns a single uint256, i.e. exactly one 32-byte word.
const erc20BalanceReturnDataLen = 32

// ProcessERC20BalanceResult decodes a balanceOf sub-call of a multicall. A
// sub-call that "succeeded" with anything but one uint256 word is reported as
// an error rather than a zero balance: Multicall3 marks a call to an address
// without code as successful with empty return data, which is what a node that
// lacks the contract's state (a lagging or misrouted RPC upstream) answers.
// Decoding that as 0 would replace a real balance with a false zero.
func ProcessERC20BalanceResult(result multicall3.IMulticall3Result) (*big.Int, error) {
	if !result.Success {
		return nil, errors.New(string(result.ReturnData))
	}
	if len(result.ReturnData) != erc20BalanceReturnDataLen {
		return nil, fmt.Errorf("unexpected balanceOf return data length: %d", len(result.ReturnData))
	}
	return new(big.Int).SetBytes(result.ReturnData), nil
}
